package reporter

import (
    "log"
    "time"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTPublisher struct { client mqtt.Client }

func NewMQTTPublisher(brokerURL, clientID string) *MQTTPublisher {
    opts := mqtt.NewClientOptions().AddBroker(brokerURL).SetClientID(clientID)
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("MQTT 连接失败: %v", token.Error())
    }
    return &MQTTPublisher{client: client}
}

func (p *MQTTPublisher) Publish(topic, payload string) {
    p.client.Publish(topic, 1, false, payload)
    log.Printf("MQTT 发送: %s -> %s", topic, payload)
}

func (p *MQTTPublisher) Disconnect() {
    p.client.Disconnect(250)
    time.Sleep(500 * time.Millisecond)
}
