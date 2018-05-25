package mqtt

import (
	"github.com/eclipse/paho.mqtt.golang"
    "../config"
    "../logging"
    "os"
)

// Quality-of-Service (at least once) for MQTT messages
const QOS = 1

type message struct {
    topic string
    msg   string
    retained bool
}

type subscriber struct {
    topic string
    subscribeCallback func()
}

type MQTTClient struct {
    mqtt mqtt.Client
    subscribers []subscriber
    toPublish []message
}

func (client *MQTTClient) Publish(topic string, msg string, retained bool) {

    if client.mqtt.IsConnected(){
        go func() {
            token := client.mqtt.Publish(topic, QOS, retained, msg)
            if token.Wait() && token.Error() == nil {
                //logging.Log.Debug("Published message to:", topic)
            } else {
                logging.Log.Error("Failed to publish message to: ", topic)
            }
        }()
    } else {
        client.toPublish = append(client.toPublish, message{topic: topic, msg:msg, retained:retained})
    }

}

func (client *MQTTClient) Subscribe(topic string, handler func([]byte)) {

    callback := func() {
        token := client.mqtt.Subscribe(topic, QOS, func(c mqtt.Client, m mqtt.Message) {
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

    // if broker connected then subscribe immediately
    if client.mqtt.IsConnected() {
        sub.subscribeCallback()
    } else {
        client.subscribers = append(client.subscribers, sub)
    }
}

func (client *MQTTClient) Unsubscribe(topic string){
    token := client.mqtt.Unsubscribe(topic)
    if token.Wait() && token.Error() == nil {
        logging.Log.Notice("Unsubscribed from: " + topic)
    } else {
        logging.Log.Notice("Failed to unsubscribe from: ", topic)
    }
}

// called when connection to mqtt broker is lost
func (client *MQTTClient) connectionLostHandler(c mqtt.Client, e error) {
    logging.Log.Error("Connection to MQTT Broker lost. Good Bye!")
    os.Exit(1)
}

// called when connection to mqtt broker established
func (client *MQTTClient) onConnectHandler(c mqtt.Client) {
    logging.Log.Notice("Connected to MQTT Broker.")
    for _, sub := range client.subscribers {
        sub.subscribeCallback()
    }
    for _, msg := range client.toPublish {
        client.Publish(msg.topic,msg.msg,msg.retained)
        logging.Log.Debugf("Published %v to %v", msg.msg, msg.topic)
    }
    client.toPublish = nil
}

func NewMQTTClient() *MQTTClient {

    conf := config.GetConfig()
    client := &MQTTClient{}

    // mqtt options
	opts := mqtt.NewClientOptions()
    opts.AddBroker(conf.BrokerAddress)
    opts.SetClientID(conf.MQTTClientName)
    opts.SetCleanSession(true)
    opts.SetAutoReconnect(false)

    logging.Log.Notice("Connecting to MQTT Broker at " + conf.BrokerAddress + " as " + conf.MQTTClientName)

    // register event handlers
    opts.SetConnectionLostHandler(client.connectionLostHandler)
    opts.SetOnConnectHandler(client.onConnectHandler)

    client.mqtt = mqtt.NewClient(opts)
    connectToken := client.mqtt.Connect()
    if connectToken.Wait() && connectToken.Error() != nil {
        logging.Log.Error("Cannot connect to MQTT Broker.", connectToken.Error())
        os.Exit(1)
    }
    return client
}