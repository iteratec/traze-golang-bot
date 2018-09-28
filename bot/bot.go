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
    "math"
)

const BROKER_GRACE_TIME = 1
const NOT_ON_GRID_TRESHOLD = 1

type Bot struct {
    mqttClient       *mqtt.MQTTClient
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
    victim           model.Bike
    hasVictim        bool
    player           []model.Player
}

func NewBot() *Bot {

    conf := config.GetConfig()
    conf.MQTTClientName = mqtt.RandStringBytesMaskImprSrc(20)
    client := mqtt.NewMQTTClient()

    bot := &Bot{mqttClient: client, ready:false, spawned:false, hasVictim:false}

    // Subscribe to grid topic
    client.Subscribe("traze/"+conf.GameInstance+"/grid", func(msg []byte) {
        var grid model.Grid
        if err := json.Unmarshal(msg, &grid); err != nil {
            logging.Log.Error("Cannot parse grid message", string(msg))
        } else {
            bot.handleGridUpdate(grid)
        }
    })

    // Subscribe to player topic
    playerTopic := "traze/"+conf.GameInstance+"/player/"+conf.MQTTClientName
    client.Subscribe(playerTopic, func(bytes []byte) {
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
        client.Unsubscribe(playerTopic)
    })

    // Subscribe to players topic
    playersTopic := "traze/"+conf.GameInstance+"/players"
    client.Subscribe(playersTopic, func(bytes []byte) {
        var player []model.Player
        if err := json.Unmarshal(bytes, &player); err != nil {
            logging.Log.Error("Cannot parse players message", string(bytes))
        } else {
            bot.player = player
        }
    })

    // Request Join. But give mqttClient time to subscribe to player topic
    time.Sleep(BROKER_GRACE_TIME*time.Second)
    joinRequest, _ := json.Marshal(model.JoinMessage{Name:conf.NickName, MQTTClientName:conf.MQTTClientName})
    logging.Log.Debug("Sending join msg to mqtt client: ", string(joinRequest))
    client.Publish("traze/"+conf.GameInstance+"/join", string(joinRequest),false)


    return bot
}

func (bot *Bot) step() {
    if len(bot.grid.Bikes) == 1 {
        bot.hasVictim = false
        bot.wallHug()
    } else {
        componentEnemies := bot.enemiesInComponent()
        if len(componentEnemies) == 0 {
            bot.hasVictim = false
            bot.wallHug() // alone in component
        } else { // enemy in component
            if bot.hasVictim && bot.contains(componentEnemies, bot.victim.PlayerId) {
                logging.Log.Infof("Attacking victim %v", bot.victim.PlayerId)
                bot.maxArea()
            } else {
                minDist := math.MaxInt32
                for  _, bike :=  range componentEnemies {
                    if bike.Distance < minDist {
                        minDist = bike.Distance
                        bot.victim = bike
                    }
                }
                logging.Log.Infof("Attacking closest enemy. Dist %v Pos %v Id %v", bot.victim.Distance, bot.victim.CurrentLocation, bot.victim.PlayerId)
                bot.hasVictim = true
                bot.maxArea()
            }
        }
    }
}

// Side effect: updates victims location
func (bot *Bot) contains(list []model.Bike, wantedId int) bool{
    for _,bike := range list {
        if bike.PlayerId == wantedId {
            bot.victim.CurrentLocation = bike.CurrentLocation
            return true
        }
    }
    return false
}

func (bot *Bot) handleGridUpdate(grid model.Grid){

    bot.grid = grid

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

// map[enemyId]distance
func (bot *Bot) enemiesInComponent() []model.Bike {

    // Make copy of grid
    var cols []model.Col
    for _, srcCol := range bot.grid.Tiles {
        col := make([]int, len(srcCol))
        copy(col, srcCol)
        cols = append(cols, col)
    }
    grid := model.Grid{Height:bot.grid.Height, Width:bot.grid.Width, Tiles: cols}

    gains := []model.Location{bot.pos}
    var enemies []model.Bike
    for { // rounds; break if no gains anymore
        lastGains := make([]model.Location, len(gains))
        copy(lastGains,gains)
        gains = nil
        for _, lastGain := range lastGains {
            for _, gain := range freeNeighbors(lastGain, grid) {
                for _,bike := range bot.grid.Bikes {
                    if bike.PlayerId != bot.playerId && ! bot.hasMyNick(bike.PlayerId) {
                        for _, enemyNeighbor := range freeNeighbors(bike.CurrentLocation,bot.grid) {
                            if enemyNeighbor == gain {
                                bike.Distance = manhattanDistance(bot.pos, bike.CurrentLocation)
                                enemies = append(enemies,bike)
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
    return enemies
}

func (bot *Bot) hasMyNick(playId int) bool {
    conf := config.GetConfig()
    for _, player := range bot.player {
        if player.Id == playId && player.Id != bot.playerId && player.Name == conf.NickName {
            return true
        }
    }
    return false
}

func (bot *Bot) wallHug()  {
    //logging.Log.Debugf("Pos: %v Dir %v. Feel like wall hugging...", bot.pos, bot.direction)
    rightTurn := turnRight(bot.direction)
    if _, ok := bot.freeNeighborsMap[rightTurn]; ok {
        if isFree(rightLeftLocation(bot.direction,bot.pos),bot.grid) {
            bot.steer(rightTurn)
        } else {
            bot.maxArea()
        }
    } else if _, ok := bot.freeNeighborsMap[bot.direction]; ok {
        if isFree(straightLeftLocation(bot.direction,bot.pos),bot.grid) {
            bot.steer(bot.direction)
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
    bot.mqttClient.Publish(bot.steerTopic,string(steerCmd),false)
    bot.direction = dir
}