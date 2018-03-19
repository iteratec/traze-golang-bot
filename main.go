package main

import (
    "./logging"
    "./bot"
)

func main() {

    logging.InitLogging()
    logging.Log.Notice("This is a traze-go-bot. Go beat me :)")

    bot.NewBot()
    //bot.NewDebug()
    select {
        // Block forever
    }
}