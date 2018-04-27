package bot

import "math"
import "../model"
import "../logging"

const WEIGHT_MY_TILES = 10000000
const WEIGHT_ENEMY_TILES = -100000
const WEIGHT_MY_ROUNDS  = -1
const DANGER_SCORE = 0

func (bot *Bot) maxArea() {

    max := math.MinInt32
    nextMove := bot.direction
    for dir, loc := range bot.freeNeighborsMap {
        score := bot.computeScore(loc)
        //logging.Log.Debugf("Move %v to %v has score %v", dir, loc, score)
        if score > max {
            max = score
            nextMove = dir
        }
    }

    logging.Log.Debugf("Pos: %v Dir: %v Possible moves: %v Going %v (%v)", bot.pos, bot.direction, bot.freeNeighborsMap, nextMove, max)
    //logging.Log.Debug(bot.grid.Tiles)
    bot.steer(nextMove)
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

    var myLastGains []model.Location
    var hisLastGains []model.Location

    round := 0
    // Mark my initial step
    myLastGains = []model.Location{{loc[0],loc[1]}}
    grid.Tiles[loc[0]][loc[1]] = bot.playerId
    numMyTiles := 1

    // Mark his position
    hisLastGains = []model.Location{bot.victim.CurrentLocation}
    numEnemyTiles := 1

    round = 1
    sumEnemyDistance := 0 // they should take long to reach their max area
    sumMyDistance := 0
    var areaGrowing bool
    var gains []model.Location
    for { // rounds; break if no gains anymore

        areaGrowing = false
        // Grow my area
        gains = nil
        for _, lastGain := range myLastGains {
            for _, gain := range freeNeighbors(lastGain, grid) {
                grid.Tiles[gain[0]][gain[1]] = bot.playerId
                gains = append(gains, gain)
            }
        }
        if len(gains) > 0{
            myLastGains = gains
            areaGrowing = true
            sumMyDistance += round
            numMyTiles += len(gains)
        }
        // Grow his area
        if bot.hasVictim {
            gains = nil
            for _, lastGain := range hisLastGains {
                for _, gain := range freeNeighbors(lastGain, grid) {
                    grid.Tiles[gain[0]][gain[1]] = bot.victim.PlayerId
                    gains = append(gains, gain)
                }
            }
            if len(gains) > 0{
                hisLastGains = gains
                areaGrowing = true
                sumEnemyDistance += round
                numEnemyTiles += len(gains)
            }
        }

        if !areaGrowing {
            break
        }
        round++
    }
    //logging.Log.Debug(grid.Tiles)
    return numMyTiles * WEIGHT_MY_TILES + numEnemyTiles * WEIGHT_ENEMY_TILES + sumEnemyDistance + sumMyDistance * WEIGHT_MY_ROUNDS
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