package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttLogPrefix = "MQTT -"
	interleave    = "Interleave"
	fast          = "Fast"
	MqttServer    = "ssl://lambo.ezplanet.org:8883"
	MqttClientId  = "vitocal-dev"
	VitocalTopic  = "climatico/vitocal"
)

type Message struct{}

var vitocalTopic string = "climatico/vitocal"
var mqttClient MQTT.Client

func init() {
	mqttServer := os.Getenv("MQTT_SERVER")
	if len(mqttServer) == 0 {
		mqttServer = MqttServer
	}
	mqttClientId := os.Getenv("MQTT_CLIENT_ID")
	if len(mqttClientId) == 0 {
		mqttClientId = MqttClientId
	}
	log.Printf("%s connecting to mqtt server: %s", mqttLogPrefix, mqttServer)
	tlsconfig := NewTLSConfig()
	opts := MQTT.NewClientOptions().
		AddBroker(mqttServer).
		SetClientID(mqttClientId).
		SetConnectionLostHandler(connLostHandler).
		SetTLSConfig(tlsconfig)

	mqttClient = MQTT.NewClient(opts)
	mqttConnect()
}

func SetVitocalTopic(topic string) {
	vitocalTopic = topic
}

// If the connection to the MQTT broker is lost, try to reconnect
func CheckConnection() {
	if !mqttClient.IsConnected() {
		mqttConnect()
	}
}

func (message *Message) Publish(topic string, retain bool, payload string) error {
	if token := mqttClient.Publish(topic, 1, retain, message); token.Wait() && token.Error() != nil {
		return token.Error()
	} else {
		return nil
	}
}

func Publish(topic string, retain bool, payload string) error {
	CheckConnection()
	if token := mqttClient.Publish(topic, 1, retain, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	} else {
		return nil
	}
}

/*** PRIVATE FUNCTIONS ***/

func mqttConnect() {
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("%s could not get connection with broker: %v", mqttLogPrefix, token.Error())
	}
}

func connLostHandler(c MQTT.Client, err error) {
	log.Printf("%s connection to broker was lost, reason: %v", mqttLogPrefix, err)
}

func NewTLSConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		//Certificates: []tls.Certificate{cert},
	}
}
