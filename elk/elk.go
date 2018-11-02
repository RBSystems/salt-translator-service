package elk

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"net/http"
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

type event struct {
	Building  string                 `json:"building"`
	Room      string                 `json:"room"`
	Cause     string                 `json:"cause"`
	Category  string                 `json:"category"`
	Hostname  string                 `json:"hostname"`
	HostType  string                 `json:"hosttype"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data",omitempty`
}

//Publish to start the listener for salt events that will then send them to the ELK stack
func Publish(events chan string, done chan bool) {

	log.L.Debugf("Publishing to ELK...")

	address := os.Getenv("ELASTIC_API_EVENTS")

	log.L.Debugf("Writing events to: %s", address)

	for {
		select {
		case event := <-events:
			send(event, address)

		case <-done:
			return
		}
	}
}

func send(event string, address string) {

	log.L.Debugf("Logging event: %v", event)

	apiEvent, myNerr := translate(event)
	if myNerr != nil {
		log.L.Debugf("Error translating event: %s: %s", event, myNerr.Error())
		return
	}

	payload, err := json.Marshal(apiEvent)
	if err != nil {
		log.L.Debugf("Error marshalling event: %v: %s", apiEvent, nerr.Translate(err).Addf("Error marshaling elk event payload to json"))
		return
	}

	response, err := http.Post(address, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.L.Debugf("Error posting event to ELK: %v: %s", apiEvent, nerr.Translate(err).Addf("Error during elk post"))
		return
	}

	value, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.L.Debugf("Error reading response: %s", nerr.Translate(err).Addf("Error during response body read"))
	}

	log.L.Debugf("Response: %s", value)

}
