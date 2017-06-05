package elk

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/byuoitav/salt-translator-service/salt"
)

type Event struct {
	Building  string                 `json:"building"`
	Room      string                 `json:"room"`
	Cause     string                 `json:"cause"`
	Category  string                 `json:"category"`
	Hostname  string                 `json:"hostname"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

var DONE bool

func Publish(events chan salt.Event, done chan bool, control *sync.WaitGroup) {

	log.Printf("Publishing to ELK...")

	var once sync.Once

	for {
		select {
		case <-done:

			log.Printf("Recieved terminate signal. Terminating ELK processes...")
			DONE = true
			control.Done()
			return

		default:
			once.Do(func() { go publishElk(events) })
		}

	}
}

func publishElk(events chan salt.Event) {

	address := os.Getenv("ELASTIC_API_EVENTS")
	log.Printf("Writing events to: %s", address)

	for {
		select {
		case event := <-events:
			go send(event, address)
		}
	}
}

func send(event salt.Event, address string) {

	if DONE {
		return
	}

	log.Printf("Logging event: %v", event)
	log.Printf("Data: %v", event.Data)

	apiEvent, err := translate(event)
	if err != nil {
		log.Printf("Error translating event: %s: %s", event.Tag, err.Error())
		return
	}

	payload, err := json.Marshal(apiEvent)
	if err != nil {
		log.Printf("Error marshalling event: %v: %s", apiEvent, err.Error())
		return
	}

	response, err := http.Post(address, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error writing event: %v: %s", apiEvent, err.Error())
		return
	}

	value, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err.Error())
	}

	log.Printf("Response: %s", value)

}
