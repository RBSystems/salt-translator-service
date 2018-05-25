package salt

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"sync"
)

func Listen(events chan Event, done chan bool, control *sync.WaitGroup) {

	log.Printf("Listening to SALT...")

	err := login()
	if err != nil {
		log.Printf("Error logging into salt. Terminating...")
		control.Done()
		return
	}

	terminate := make(chan bool)
	go ListenSalt(events, terminate)

	<-done
	terminate <- true

	log.Printf("Received terminate signal. Terminating SALT process...")
	connection.Response.Body.Close()
	control.Done()
}

func ListenSalt(events chan Event, terminate chan bool) {

	log.Printf("Reading salt events...")

	for {

		select {

		case <-terminate:
			return

		default:

			line, err := connection.Reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading event %s. Terminating", err.Error())
				break
			} else {
				if strings.Contains(line, "retry") {
					continue
				} else if strings.Contains(line, "tag") {

					line2, err := connection.Reader.ReadString('\n')
					if err != nil {
						log.Fatal(err)
					}

					err = ReadAndWriteEvent(line2, events)
					if err != nil {
						log.Printf("Error reading event: %s", err.Error())
					}

				} else if len(line) < 1 {
					continue
				}
			}
		}
	}

	return
}

func ReadAndWriteEvent(line string, events chan Event) error {

	if strings.Contains(line, "data") {

		jsonString := line[5:]

		var event Event
		err := json.Unmarshal([]byte(jsonString), &event)
		if err != nil {
			log.Fatal("Error unmarshalling event" + err.Error())
		}

		ok, err := Filter(event)
		if err != nil {
			log.Printf("Error evaluating event: %s", err.Error())
		}

		if ok {
			log.Printf("Writing event to channel: %v", event)
			events <- event
		}

	}

	return nil
}

func Filter(event Event) (bool, error) {
	log.Printf("Evaluating event: %s", event.Tag)

	auth := regexp.MustCompile(`\/auth`)
	if auth.MatchString(event.Tag) {
		log.Printf("Filtering out salt authorization event: %s", event.Tag)
		return false, nil
	}
	staging := regexp.MustCompile(`STAGE-STG`)
	if staging.MatchString(event.Tag) {
		log.Printf("Filtering out events from Pis in staging: %s", event.Tag)
		return false, nil
	}

	return true, nil
}
