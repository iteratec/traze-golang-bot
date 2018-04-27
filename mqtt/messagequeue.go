package mqtt

import (
    "time"
	"github.com/eclipse/paho.mqtt.golang"
    "../config"
    "../logging"
)

// Quality-of-Service (at least once) for MQTT messages
const _QOS = 1

// in milliseconds
const _MAX_RECONNECT_INTERVAL = 30000
const _INIT_RECONNECT_INTERVAL = 500

// MessageHandler on receiving message from subscribed topic
type MessageHandler func([]byte)

type subscriber struct {
    topic string
    subscribeCallback func()
}

type MessageQueue struct {
    //
    client mqtt.Client
    //
    subscribers []subscriber
    // time to wait before reconnect
    reconnectInterval time.Duration
}

func (mq *MessageQueue) Publish(topic string, msg string, retained bool) {
    go func() {
        token := mq.client.Publish(topic, _QOS, retained, msg)
        token.Wait()
        if token.Wait() && token.Error() == nil {
            //logging.Log.Notice("Published message to:", topic)
        } else {
            logging.Log.Error("Failed to publish message to: ", topic)
        }
    }()
}

func (mq *MessageQueue) Subscribe(topic string, handler func([]byte)) {

    callback := func() {
        token := mq.client.Subscribe(topic, _QOS, func(c mqtt.Client, m mqtt.Message) {
            handler(m.Payload())
        })

        if token.Wait() && token.Error() == nil {
            logging.Log.Notice("Subscribed to: " + topic)
        } else {
            logging.Log.Notice("Failed to subscribe to: ", topic)
        }
    }

    // save the subscribers for later
    sub := subscriber{topic:topic, subscribeCallback:callback}
    mq.subscribers = append(mq.subscribers, sub)

    // if broker connected then subscribe immediately
    if mq.client.IsConnected() {
        go sub.subscribeCallback()
    }
}

func (mq *MessageQueue) Unsubscribe(topic string){
    token := mq.client.Unsubscribe(topic)
    if token.Wait() && token.Error() == nil {
        logging.Log.Notice("Unsubscribed from: " + topic)
    } else {
        logging.Log.Notice("Failed to unsubscribe from: ", topic)
    }
}

// called when connection to mqtt broker is lost
func (mq *MessageQueue) connectionLostHandler(c mqtt.Client, e error) {
    logging.Log.Error("Connection to MQTT Broker lost. Trying to reconnect ...")
}

// called when connection to mqtt broker established
func (mq *MessageQueue) onConnectHandler(c mqtt.Client) {
    logging.Log.Notice("Connected to MQTT Broker.")
    mq.reconnectInterval = _INIT_RECONNECT_INTERVAL
    for i := 0; i < len(mq.subscribers); i++ {
        mq.subscribers[i].subscribeCallback()
    }
}

// initialize a new message queue
func NewMessageQueue() *MessageQueue {

    conf := config.GetConfig()
    mq := &MessageQueue{}

    // initialize subscribers slice
    mq.reconnectInterval = _INIT_RECONNECT_INTERVAL
    mq.subscribers = make([]subscriber, 0)

    // mqtt options
	opts := mqtt.NewClientOptions()
    opts.AddBroker(conf.BrokerAddress)
    opts.SetClientID(conf.MQTTClientName)
    opts.SetCleanSession(true)
    opts.SetAutoReconnect(true)

    logging.Log.Notice("Connecting to MQTT Broker at " + conf.BrokerAddress + " as " + conf.MQTTClientName)

    // register event handlers
    opts.SetConnectionLostHandler(mq.connectionLostHandler)
    opts.SetOnConnectHandler(mq.onConnectHandler)

    mq.client = mqtt.NewClient(opts)

    // connect callback with retry
    var connectCallback func()
    connectCallback = func () {
        connectToken := mq.client.Connect()
        if connectToken.Wait() && connectToken.Error() != nil {
            logging.Log.Error("Cannot connect to MQTT Broker.", connectToken.Error())
            logging.Log.Noticef("Trying to reconnect to MQTT Broker after %d ms\n", mq.reconnectInterval)

            time.Sleep(mq.reconnectInterval * time.Millisecond)
            go connectCallback()

            // increase reconnect interval by 1.2 but should not exceed _MAX_RECONNECT_INTERVAL
            mq.reconnectInterval = time.Duration(float64(mq.reconnectInterval) * 1.2)
            if mq.reconnectInterval > _MAX_RECONNECT_INTERVAL {
                mq.reconnectInterval = _MAX_RECONNECT_INTERVAL
            }
        }
    }

    go connectCallback()

    return mq
}