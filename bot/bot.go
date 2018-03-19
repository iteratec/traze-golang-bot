package bot

import (
    "../mqtt"
    "encoding/json"
    "../model"
    "../config"
    "../logging"
    "strconv"
    "time"
    "math/rand"
    "os"
    "math"
)

const BROKER_GRACE_TIME = 2
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

    rand.Seed(time.Now().Unix())
    conf := config.GetConfig()
    mq := mqtt.NewMessageQueue()

    bot := &Bot{mq:mq, ready:false}

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

    // Request Join. But give mq time to connect to server
    time.Sleep(BROKER_GRACE_TIME*time.Second)
    joinRequest, _ := json.Marshal(model.JoinMessage{Name:conf.ClientName, MQTTClientName:conf.ClientName})
    mq.Publish("traze/"+conf.GameInstance+"/join", string(joinRequest),false)
    logging.Log.Debug("Joing msg: ", string(joinRequest))

    return bot
}

func NewDebug() {
    bot := &Bot{
        playerId: 1,
        direction: model.N,
        pos: model.Location{1,3},
        grid:model.Grid{
            Width:5,
            Height:4,
            Bikes: []model.Bike{
                {
                    PlayerId:1,
                    CurrentLocation:model.Location{1,3},
                },
                {
                    PlayerId:2,
                    CurrentLocation:model.Location{3,2},
                },
            },
            Tiles:[]model.Col{
                {2,0,0,0},
                {0,2,2,1},
                {0,0,2,0},
                {0,0,2,0},
                {0,0,0,0},
            },
        },
    }
    bot.freeNeighborsMap = freeNeighborsMap(model.Location{1,3},bot.grid)

    bot.step()
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

func (bot *Bot) maxArea() {

    max := math.MinInt32
    nextMove := bot.direction
    for dir, loc := range bot.freeNeighborsMap {
        score := bot.computeScore(loc)
        logging.Log.Debugf("Move %v to %v has score %v", dir, loc, score)
        if score > max {
            max = score
            nextMove = dir
        }
    }

    logging.Log.Debugf("Pos: %v Dir: %v Possible moves: %v Going %v (%v)", bot.pos, bot.direction, bot.freeNeighborsMap, nextMove, max)
    //logging.Log.Debug(bot.grid.Tiles)
    bot.steer(nextMove)
}

func (bot *Bot) isDanger(loc model.Location, grid model.Grid) bool {
    var freeNeighborsEnemies []model.Location
    for _, bike := range bot.grid.Bikes {
        if bike.PlayerId != bot.playerId {
            freeNeighborsEnemies = append(freeNeighborsEnemies, freeNeighbors(bike.CurrentLocation, grid)...)
        }
    }
    for _, danger := range freeNeighborsEnemies {
        if loc == danger {
            return true
        }
    }
    return false
}

func (bot *Bot) computeScore(loc model.Location) int {

    if bot.isDanger(loc, bot.grid) {
        logging.Log.Debug("Danger: ", loc)
        return DANGER_SCORE
    }

    // Make copy of grid
    var cols []model.Col
    for _, srcCol := range bot.grid.Tiles {
        col := make([]int, len(srcCol))
        copy(col, srcCol)
        cols = append(cols, col)
    }
    grid := model.Grid{Height:bot.grid.Height, Width:bot.grid.Width, Tiles: cols}

    // map[playerId]map[round][]model.Location
    playerLastGains := make(map[int]map[int][]model.Location)

    round := 0
    // Mark my initial step
    playerLastGains[bot.playerId] = make(map[int][]model.Location)
    playerLastGains[bot.playerId][round] = []model.Location{{loc[0],loc[1]}}
    grid.Tiles[loc[0]][loc[1]] = bot.playerId

    // Let enemies gain initial area
    for _, bike := range bot.grid.Bikes {
        if bike.PlayerId != bot.playerId {
            var gains []model.Location
            for _, gain := range freeNeighbors(bike.CurrentLocation, grid) {
                grid.Tiles[gain[0]][gain[1]] = bike.PlayerId
                gains = append(gains, gain)
            }
            playerLastGains[bike.PlayerId] = make(map[int][]model.Location)
            playerLastGains[bike.PlayerId][round] = gains
        }
    }

    round = 1
    sumEnemyDistance := 0 // they should take long to reach their max area
    sumMyDistance := 0
    var areaGrowing bool
    for { // rounds; break if no gains anymore
        areaGrowing = false
        for _, bike := range bot.grid.Bikes {
            var gains []model.Location
            for _, lastGain := range playerLastGains[bike.PlayerId][round-1] {
                for _, gain := range freeNeighbors(lastGain, grid) {
                    grid.Tiles[gain[0]][gain[1]] = bike.PlayerId
                    gains = append(gains, gain)
                }
            }
            if len(gains) > 0{
                playerLastGains[bike.PlayerId][round] = gains
                areaGrowing = true
                if bike.PlayerId != bot.playerId {
                    sumEnemyDistance += round
                } else {
                    sumMyDistance += round
                }
            }
        }
        if !areaGrowing {
            break
        }
        round++
    }

    numMyTiles := 0
    numEnemyTiles := 0
    for _, col := range grid.Tiles {
        for _, tile := range col{
            if tile == bot.playerId {
                numMyTiles++
            } else if tile != 0 {
                numEnemyTiles++
            }
        }
    }

    return numMyTiles * WEIGHT_MY_TILES + numEnemyTiles * WEIGHT_ENEMY_TILES + sumEnemyDistance + sumMyDistance * WEIGHT_MY_ROUNDS
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
    rightTurn := bot.turnRight(bot.direction)
    if _, ok := bot.freeNeighborsMap[rightTurn]; ok {
        if isFree(bot.rightLeftLocation(),bot.grid) {
            bot.steer(rightTurn)
        } else {
            bot.maxArea()
        }
    } else if _, ok := bot.freeNeighborsMap[bot.direction]; ok {
        if isFree(bot.straightLeftLocation(),bot.grid) {
            // keep going straight
        } else {
            bot.maxArea()
        }
    } else {
        bot.steer(bot.turnLeft(bot.direction))
    }
}


func (bot *Bot) steer(dir model.Cardinal) {
    steerCmd, _ := json.Marshal(model.SteerCommand{Course: model.CardinalStringMap[dir], PlayerToken: bot.token})
    //logging.Log.Debug("Going to steer: ", string(steerCmd))
    bot.mq.Publish(bot.steerTopic,string(steerCmd),false)
    bot.direction = dir
}


func (bot *Bot) turnRight(dir model.Cardinal) model.Cardinal {
    if dir == model.W {
        return model.N
    } else {
        return dir + 1
    }
}

func (bot *Bot) turnLeft(dir model.Cardinal) model.Cardinal {
    if dir == model.N {
        return model.W
    } else {
        return dir - 1
    }
}

func (bot *Bot) straightLeftLocation() model.Location {
    if bot.direction == model.N {
        return model.Location{bot.pos[0]-1,bot.pos[1]+1}
    } else if bot.direction == model.E {
        return model.Location{bot.pos[0]+1,bot.pos[1]+1}
    } else if bot.direction == model.S {
        return model.Location{bot.pos[0]+1, bot.pos[1]-1}
    } else {
        return model.Location{bot.pos[0]-1,bot.pos[1]-1}
    }
}

func (bot *Bot) rightLeftLocation() model.Location {
    if bot.direction == model.N {
        return model.Location{bot.pos[0]+1,bot.pos[1]+1}
    } else if bot.direction == model.E {
        return model.Location{bot.pos[0]+1,bot.pos[1]-1}
    } else if bot.direction == model.S {
        return model.Location{bot.pos[0]-1, bot.pos[1]-1}
    } else {
        return model.Location{bot.pos[0]-1,bot.pos[1]+1}
    }
}

func (bot *Bot) rightLocation() model.Location {
    if bot.direction == model.N {
        return model.Location{bot.pos[0]+1,bot.pos[1]}
    } else if bot.direction == model.E {
        return model.Location{bot.pos[0],bot.pos[1]-1}
    } else if bot.direction == model.S {
        return model.Location{bot.pos[0]-1, bot.pos[1]}
    } else {
        return model.Location{bot.pos[0],bot.pos[1]+1}
    }
}

func (bot *Bot) leftLocation() model.Location {
    if bot.direction == model.N {
        return model.Location{bot.pos[0]-1,bot.pos[1]}
    } else if bot.direction == model.E {
        return model.Location{bot.pos[0],bot.pos[1]+1}
    } else if bot.direction == model.S {
        return model.Location{bot.pos[0]+1, bot.pos[1]}
    } else {
        return model.Location{bot.pos[0],bot.pos[1]-1}
    }
}

func freeNeighborsMap(pos model.Location, grid model.Grid) map[model.Cardinal]model.Location {
    dic := make(map[model.Cardinal]model.Location)
    var nextPos model.Location

    nextPos = nextPosition(pos,model.N)
    if isFree(nextPos, grid) {
        dic[model.N] = nextPos
    }

    nextPos = nextPosition(pos,model.E)
    if isFree(nextPos, grid) {
        dic[model.E] = nextPos
    }

    nextPos = nextPosition(pos,model.S)
    if isFree(nextPos, grid) {
        dic[model.S] = nextPos
    }

    nextPos = nextPosition(pos,model.W)
    if isFree(nextPos, grid) {
        dic[model.W] = nextPos
    }
    return dic
}

func freeNeighbors(pos model.Location, grid model.Grid) []model.Location {
    var free []model.Location
    if isFree(model.Location{pos[0]+1,pos[1]}, grid) {
        free = append(free, model.Location{pos[0]+1,pos[1]})
    }
    if isFree(model.Location{pos[0],pos[1]+1}, grid) {
        free = append(free, model.Location{pos[0],pos[1]+1})
    }
    if isFree(model.Location{pos[0],pos[1]-1}, grid) {
        free = append(free, model.Location{pos[0],pos[1]-1})
    }
    if isFree(model.Location{pos[0]-1,pos[1]}, grid) {
        free = append(free, model.Location{pos[0]-1,pos[1]})
    }
    return free
}

func isFree(loc model.Location, grid model.Grid) bool {
    return !isBorder(loc,grid) && !isTrail(loc,grid)
}

func isBorder(loc model.Location, grid model.Grid) bool {
    return loc[0] >= grid.Width || loc[0] < 0 || loc[1] < 0 || loc[1] >= grid.Height
}

func isTrail(loc model.Location, grid model.Grid) bool {
    return grid.Tiles[loc[0]][loc[1]] != 0
}

func nextPosition(pos model.Location, dir model.Cardinal) model.Location {
    if dir == model.N {
        return model.Location{pos[0],pos[1]+1}
    } else if dir == model.E {
        return model.Location{pos[0]+1, pos[1]}
    } else if dir == model.S {
        return model.Location{pos[0],pos[1]-1}
    } else {
        return model.Location{pos[0]-1,pos[1]}
    }
}