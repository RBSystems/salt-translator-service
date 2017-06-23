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

	var client, listener sync.Once

	for {

		select {
		case <-done:

			log.Printf("Received terminate signal. Terminating SALT process...")
			connection.Response.Body.Close()
			control.Done()
			return

		default:

			client.Do(login)
			listener.Do(func() { go listenSalt(events) })

		}
	}
}

func listenSalt(events chan Event) {

	log.Printf("Reading salt events...")

	for {

		if connection.Response.Close {
			log.Printf("Detected closed salt connection. Terminating process...")
			break
		}

		line, err := connection.Reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading event" + err.Error())
		} else {
			if strings.Contains(line, "retry") {
				continue
			} else if strings.Contains(line, "tag") {

				line2, err := connection.Reader.ReadString('\n')
				if err != nil {
					log.Fatal(err)
				}

				if strings.Contains(line2, "data") {

					jsonString := line2[5:]

					var event Event
					err := json.Unmarshal([]byte(jsonString), &event)
					if err != nil {
						log.Fatal("Error unmarshalling event" + err.Error())
					}

					ok, err := filter(event)
					if err != nil {
						log.Printf("Error evaluating event: %s", err.Error())
					}

					if ok {
						log.Printf("Writing event to channel: %v", event)
						events <- event
					}

				}

			} else if len(line) < 1 {
				continue
			}
		}
	}
	return
}

func filter(event Event) (bool, error) {
	log.Printf("Evaluating event: %s", event.Tag)

	auth := regexp.MustCompile(`\/auth`)
	if auth.MatchString(event.Tag) {
		log.Printf("Filtering out salt authorization event: %s", event.Tag)
		return false, nil
	}

	return true, nil
}
