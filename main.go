package main

import (
    "./logging"
    "./bot"
)

func main() {

    logging.InitLogging()
    logging.Log.Notice("This is traze-go-bot. Go beat me!")

    bot.NewBot()
    //bot.NewDebug()
    select {
        // Block forever
    }
}