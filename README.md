# Traze Golang Bot

* A bot for the MQTT network based traze challange.
* See https://github.com/iteratec/traze 
* Implements an "area maximizing" bot.
* Inspired by the idea from Pascal van Kooten (see Acknowledgements).

## Improvements to Pascal's idea:
* Computes connected components.
* Attacks a single victim within the own component.
* Falls back to wall hugging when alone in component.

## Build
`cd traze-go-bot && go build`

## Usage
`MQTT_BROKER_ADDRESS="tcp://traze.iteratec.de:1883" GAME_INSTANCE="1" CLIENT_NAME="MyClientName" ./traze-go-bot`

## Ackknowledgements
* Pascal van Kooten (https://kootenpv.github.io/2016-09-07-ai-challenge-in-78-lines)
