package main

import (
    "./logging"
    "./bot"
    "time"
)

const BROKER_GRACE_TIME = 1

func main() {

    logging.InitLogging()
    logging.Log.Notice("This is traze-go-bot. Go beat me!")

    bot := bot.NewBot()
    time.Sleep(BROKER_GRACE_TIME*time.Second)
    bot.JoinGame()

    select {
        // Block forever
    }
}