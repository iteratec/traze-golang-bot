package bot

import "../model"

func NewDebug() {
    enemy := model.Bike{
        PlayerId: 2,
        CurrentLocation:model.Location{3,0},
    }
    me := model.Bike{
        PlayerId:        1,
        CurrentLocation: model.Location{3,2},
    }

    bot := &Bot{
        playerId:  me.PlayerId,
        direction: model.S,
        pos:       me.CurrentLocation,
        grid: model.Grid{
            Width: 6,
            Height: 6,
            Bikes: []model.Bike{me,enemy},
            Tiles: []model.Col{
                {2,0,0,0,0,0},
                {2,0,0,0,0,0},
                {2,0,0,0,0,0},
                {2,0,1,1,1,0},
                {0,0,0,0,0,0},
                {0,0,0,0,0,0},
            },
        },
        victim: enemy,
        hasVictim: true,
    }
    bot.freeNeighborsMap = freeNeighborsMap(bot.pos,bot.grid)

    bot.step()
}