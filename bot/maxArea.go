package bot

import "math"
import "../model"
import "../logging"

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