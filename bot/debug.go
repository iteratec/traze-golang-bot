package bot

import "../model"

func NewDebug() {
    enemy := model.Bike{
        PlayerId: 2,
        CurrentLocation:model.Location{0,0},
    }
    me := model.Bike{
        PlayerId:        1,
        CurrentLocation: model.Location{3,1},
    }

    bot := &Bot{
        playerId:  me.PlayerId,
        //direction: model.S,
        pos:       me.CurrentLocation,
        grid: model.Grid{
            Width: 4,
            Height: 4,
            Bikes: []model.Bike{me,enemy},
            Tiles: []model.Col{
                {2,0,0,0},
                {0,0,0,0},
                {0,0,0,0},
                {0,1,0,0},
            },
        },
        victim: enemy,
        hasVictim: true,
    }
    bot.freeNeighborsMap = freeNeighborsMap(bot.pos,bot.grid)

    bot.step()
}