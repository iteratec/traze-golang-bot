package bot

import "../model"

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