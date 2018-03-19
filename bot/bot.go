package bot

import (
    "../mqtt"
    "encoding/json"
    "../model"
    "../config"
    "../logging"
    "strconv"
    "time"
    "os"
)

const BROKER_GRACE_TIME = 500
const NOT_ON_GRID_TRESHOLD = 1
const WEIGHT_MY_TILES = 10000000
const WEIGHT_ENEMY_TILES = -100000
const WEIGHT_MY_ROUNDS  = 0
const DANGER_SCORE = 0

type Bot struct {
    mq               *mqtt.MessageQueue
    token            string
    playerId         int
    steerTopic       string
    ready            bool
    spawned          bool
    direction        model.Cardinal
    pos              model.Location
    grid             model.Grid
    countNotOnGrid   int
    freeNeighborsMap map[model.Cardinal]model.Location
}

func NewBot() *Bot {

    conf := config.GetConfig()
    mq := mqtt.NewMessageQueue()

    bot := &Bot{mq:mq, ready:false, spawned:false}

    // Subscribe to grid topic
    mq.Subscribe("traze/"+conf.GameInstance+"/grid", func(msg []byte) {
        var grid model.Grid
        if err := json.Unmarshal(msg, &grid); err != nil {
            logging.Log.Error("Cannot parse grid message", string(msg))
        } else {
            bot.handleGridUpdate(grid)
        }
    })

    // Subscribe to player topic
    playerTopic := "traze/"+conf.GameInstance+"/player/"+conf.ClientName
    mq.Subscribe(playerTopic, func(bytes []byte) {
        var playerMsg model.PlayerMessage
        if err := json.Unmarshal(bytes, &playerMsg); err != nil {
            logging.Log.Error("Cannot parse player message", string(bytes))
        } else {
            bot.token = playerMsg.SecretUserToken
            bot.pos = playerMsg.Position
            bot.playerId = playerMsg.Id
            bot.steerTopic = "traze/"+conf.GameInstance+"/"+strconv.Itoa(playerMsg.Id)+"/steer"
            logging.Log.Debug("Received private msg: ", string(bytes))

            bot.spawned = true
            bot.ready = true
        }
        mq.Unsubscribe(playerTopic)
    })

    // Request Join. But give mq time to subscribe to player topic
    time.Sleep(BROKER_GRACE_TIME*time.Millisecond)
    joinRequest, _ := json.Marshal(model.JoinMessage{Name:conf.ClientName, MQTTClientName:conf.ClientName})
    mq.Publish("traze/"+conf.GameInstance+"/join", string(joinRequest),false)
    logging.Log.Debug("Joing msg: ", string(joinRequest))

    return bot
}

func (bot *Bot) step() {
    if len(bot.grid.Bikes) == 1 {
        bot.wallHug()
    } else {
        if bot.aloneInComponent() {
            bot.wallHug()
        } else {
            bot.maxArea()
        }
    }
}

func (bot *Bot) handleGridUpdate(grid model.Grid){

    bot.grid = grid

    // Fixing Server bugs
    bot.grid.Tiles = bot.grid.Tiles[:bot.grid.Width]
    for i, tile := range bot.grid.Tiles {
        bot.grid.Tiles[i] = tile[:bot.grid.Height]
    }
    for _, bike := range grid.Bikes {
        bot.grid.Tiles[bike.CurrentLocation[0]][bike.CurrentLocation[1]] = bike.PlayerId
    }

    if bot.ready {
        if bot.spawned {
            bot.freeNeighborsMap = freeNeighborsMap(bot.pos,bot.grid)
            bot.spawned = false
        } else {
            positionFound := false
            for _, bike := range bot.grid.Bikes {
                if bike.PlayerId == bot.playerId{
                    positionFound = true
                    bot.pos = bike.CurrentLocation
                    break
                }
            }
            if !positionFound {
                logging.Log.Info("Unable to read my position from grid")
                bot.countNotOnGrid++
                if bot.countNotOnGrid > NOT_ON_GRID_TRESHOLD{
                    logging.Log.Infof("More than %v in a row not on grid. God bye!", NOT_ON_GRID_TRESHOLD)
                    os.Exit(0)
                }
                return
            }
            bot.countNotOnGrid = 0
            bot.freeNeighborsMap = freeNeighborsMap(bot.pos,bot.grid)
        }

        bot.step()
    }
}

func (bot *Bot) aloneInComponent() bool {

    // Make copy of grid
    var cols []model.Col
    for _, srcCol := range bot.grid.Tiles {
        col := make([]int, len(srcCol))
        copy(col, srcCol)
        cols = append(cols, col)
    }
    grid := model.Grid{Height:bot.grid.Height, Width:bot.grid.Width, Tiles: cols}

    gains := []model.Location{bot.pos}
    for { // rounds; break if no gains anymore
        lastGains := make([]model.Location, len(gains))
        copy(lastGains,gains)
        gains = nil
        for _, lastGain := range lastGains {
            for _, gain := range freeNeighbors(lastGain, grid) {
                for _,bike := range bot.grid.Bikes {
                    if bike.PlayerId != bot.playerId {
                        for _, enemyNeighbor := range freeNeighbors(bike.CurrentLocation,bot.grid) {
                            if enemyNeighbor == gain {
                                return false
                            }
                        }
                    }
                }
                gains = append(gains, gain)
                grid.Tiles[gain[0]][gain[1]] = bot.playerId
            }
        }
        if len(gains) == 0{
            break // no gain
        }
    }
    return true
}

func (bot *Bot) wallHug()  {
    rightTurn := turnRight(bot.direction)
    if _, ok := bot.freeNeighborsMap[rightTurn]; ok {
        if isFree(rightLeftLocation(bot.direction,bot.pos),bot.grid) {
            bot.steer(rightTurn)
        } else {
            bot.maxArea()
        }
    } else if _, ok := bot.freeNeighborsMap[bot.direction]; ok {
        if isFree(straightLeftLocation(bot.direction,bot.pos),bot.grid) {
            // keep going straight
        } else {
            bot.maxArea()
        }
    } else {
        bot.steer(turnLeft(bot.direction))
    }
}

func (bot *Bot) steer(dir model.Cardinal) {
    steerCmd, _ := json.Marshal(model.SteerCommand{Course: model.CardinalStringMap[dir], PlayerToken: bot.token})
    //logging.Log.Debug("Going to steer: ", string(steerCmd))
    bot.mq.Publish(bot.steerTopic,string(steerCmd),false)
    bot.direction = dir
}