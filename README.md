# Traze Golang Bot

* A bot for the MQTT network based traze challange.
* See https://github.com/iteratec/traze .
* Implements an "area maximizing" bot.
* Inspired by the idea from Pascal van Kooten (see Acknowledgements).

## Improvements to Pascal's Idea:
* Computes connected components.
* Attacks a single victim within the own component.
* Falls back to wall hugging when alone in component.

## Run
```
docker run \
 -e MQTT_BROKER_ADDRESS="tcp://traze.iteratec.de:1883" \
 -e GAME_INSTANCE="1" \
 -e NICK_NAME="CHANGE-ME" \
  iteratec/traze-golang-bot
```
Please choose your own NICK_NAME.

## Build and Run Locally
```
git clone git@github.com:iteratec/traze-golang-bot.git
cd traze-golang-bot
docker build -t traze-golang-bot .
docker run \
 -e MQTT_BROKER_ADDRESS="tcp://traze.iteratec.de:1883" \
 -e GAME_INSTANCE="1" \
 -e NICK_NAME="CHANGE-ME" \
 traze-golang-bot
```
Please choose your own NICK_NAME.

## Ackknowledgements
* Pascal van Kooten (https://kootenpv.github.io/2016-09-07-ai-challenge-in-78-lines)

## Contributers
* Max Berndt (https://github.com/Mexx77)
