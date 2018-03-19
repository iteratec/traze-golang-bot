package bot

import "../model"

func turnRight(dir model.Cardinal) model.Cardinal {
    if dir == model.W {
        return model.N
    } else {
        return dir + 1
    }
}

func turnLeft(dir model.Cardinal) model.Cardinal {
    if dir == model.N {
        return model.W
    } else {
        return dir - 1
    }
}

func straightLeftLocation(dir model.Cardinal, pos model.Location) model.Location {
    if dir == model.N {
        return model.Location{pos[0]-1, pos[1]+1}
    } else if dir == model.E {
        return model.Location{pos[0]+1, pos[1]+1}
    } else if dir == model.S {
        return model.Location{pos[0]+1, pos[1]-1}
    } else {
        return model.Location{pos[0]-1, pos[1]-1}
    }
}

func rightLeftLocation(dir model.Cardinal, pos model.Location) model.Location {
    if dir == model.N {
        return model.Location{pos[0]+1, pos[1]+1}
    } else if dir == model.E {
        return model.Location{pos[0]+1, pos[1]-1}
    } else if dir == model.S {
        return model.Location{pos[0]-1, pos[1]-1}
    } else {
        return model.Location{pos[0]-1, pos[1]+1}
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