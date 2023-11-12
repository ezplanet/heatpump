//BSD 2-Clause License
//
//Copyright (c) 2023, Mauro Mozzarelli
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are met:
//
//1. Redistributions of source code must retain the above copyright notice, this
//list of conditions and the following disclaimer.
//
//2. Redistributions in binary form must reproduce the above copyright notice,
//this list of conditions and the following disclaimer in the documentation
//and/or other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
//AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
//IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
//FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
//DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
//SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
//CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
//OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
//OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"heatpump/base"
	"log"
)

const (
	mqttLogPrefix = "MQTT -"
	interleave    = "Interleave"
	fast          = "Fast"
)

type Message struct{}

var mqttClient MQTT.Client

func init() {
	log.Printf("%s connecting to mqtt server: %s", mqttLogPrefix, base.MqttServer)
	tlsconfig := NewTLSConfig()
	opts := MQTT.NewClientOptions().
		AddBroker(base.MqttServer).
		SetClientID(base.MqttClientId).
		SetConnectionLostHandler(connLostHandler).
		SetTLSConfig(tlsconfig)

	mqttClient = MQTT.NewClient(opts)
	mqttConnect()
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
