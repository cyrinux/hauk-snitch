package mapper

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/mdp/qrterminal"
	"github.com/tuffnerdstuff/hauk-snitch/hauk"
	"github.com/tuffnerdstuff/hauk-snitch/mqtt"
)

var keyMap = map[string]string{
	"acc": "acc",
	"alt": "alt",
	"lat": "lat",
	"lon": "lon",
	"vel": "spd",
	"tst": "time",
}

var topicSessionMap = make(map[string]hauk.Session)

// Run maps mqtt messages to hauk API calls
func Run(messages <-chan mqtt.Message, haukClient *hauk.Client) {
	for message := range messages {
		locationParams, err := createLocationParamsFromMessage(message)
		if err != nil {
			log.Printf("Message invalid, skipping: %s\n", err.Error())
			continue
		}

		sid, err := getCurrentSIDForTopic(message.Topic, haukClient)
		if err != nil {
			log.Printf("%v", err.Error())
			continue
		}
		err = haukClient.PostLocation(sid, locationParams)
		if err != nil {
			switch err.(type) {
			case *hauk.SessionExpiredError:
				log.Printf("Session for %s expired, creating new one\n", message.Topic)
				if _, err := createNewSIDForTopic(message.Topic, haukClient); err != nil {
					log.Printf("%v", err.Error())
					continue
				}
			default:
				log.Printf("Invalid location %v: %v\n", locationParams, err.Error())
				continue
			}
		}
	}
}

func createLocationParamsFromMessage(msg mqtt.Message) (url.Values, error) {
	body := msg.Body
	haukValues := url.Values{}
	if body[mqtt.ParamType] == "location" {
		for mqttKey, mqttValue := range body {
			setHaukValue(&haukValues, mqttKey, mqttValue)

		}
		return haukValues, nil
	}
	return haukValues, fmt.Errorf("Type is not location")

}

func getCurrentSIDForTopic(topic string, haukClient *hauk.Client) (string, error) {
	session, sessionExists := topicSessionMap[topic]
	if !sessionExists {
		log.Printf("New topic %s, creating session\n", topic)
		return createNewSIDForTopic(topic, haukClient)
	}
	return session.SID, nil
}

func createNewSIDForTopic(topic string, haukClient *hauk.Client) (string, error) {
	newSession, err := haukClient.CreateSession()
	if err != nil {
		return "n/a", err
	}
	topicSessionMap[topic] = newSession

	// Print QR code on terminal
	log.Printf("New session for %s: %v", topic, newSession)
	qrterminal.GenerateHalfBlock(newSession.URL, qrterminal.L, os.Stdout)

	return newSession.SID, nil
}

func setHaukValue(haukValues *url.Values, key string, value interface{}) {
	haukKey, hasMapping := keyMap[key]
	if hasMapping {
		haukValues.Add(haukKey, fmt.Sprintf("%v", value))
	}
}
