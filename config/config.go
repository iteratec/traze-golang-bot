package config

import (
    "../logging"
    "os"
)

var config *BotConfig

// env var constants
const mqttBrokerAddress =   "MQTT_BROKER_ADDRESS"
const clientName =   "CLIENT_NAME"
const gameInstance = "GAME_INSTANCE"

type BotConfig struct {
    BrokerAddress string
    ClientName    string
    GameInstance  string
}

// loads the configuration from environment variables and checks constraints
func GetConfig() *BotConfig {
    if config == nil{

        config = &BotConfig{
            BrokerAddress: getConfigString(mqttBrokerAddress, "Missing broker address."),
            ClientName:    getConfigString(clientName, "Missing client name."),
            GameInstance:  getConfigString(gameInstance, "Missing game instance."),
        }
        return config
    } else {
        return config
    }
}

// loads a single string value out of a given environment variable.
// if the variable is empty a given error message will be printed.
func getConfigString(envVar string, errorMsg string) string {
    entry := os.Getenv(envVar)
    if entry == "" {
        logging.Log.Critical(errorMsg + " Please configure " + envVar)
        os.Exit(1)
    }
    logging.Log.Debug(envVar+":", entry)
    return entry
}
