package salt

import (
	"time"

	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

//Listen starts up and listens to salt until it is done
func Listen(events chan string, done chan bool) {

	log.L.Debugf("Listening to SALT...")

	connection := createSaltConnection(done)

	log.L.Debugf("Reading salt events...")

	for {

		select {

		case <-done:
			connection.Response.Body.Close()

			return

		default:

			line, err := connection.Reader.ReadString('\n')
			if err != nil {
				log.L.Debugf("Error reading event %s. Attempting to re-establish connection", err.Error())
				connection.Response.Body.Close()
				connection = createSaltConnection(done)
			} else {
				if strings.Contains(line, "retry") {
					continue
				} else if strings.Contains(line, "tag") {

					line2, err := connection.Reader.ReadString('\n')
					if err != nil {
						log.L.Debugf("Error reading the next line of event: %v", err.Error())
						continue
					}

					nerr := readAndWriteEvent(line2, events)
					if nerr != nil {
						log.L.Debugf("Error reading event: %s", nerr.Error())
					}

				} else if len(line) < 1 {
					continue
				}
			}
		}
	}
}

func createSaltConnection(done chan bool) saltConnection {
	var connection saltConnection

	//create the connection.  If it fails, then
	//wait for 1 second and retry
	//keep retrying after 1s, 10s, 30s, 1m, 2m, and then every 5m indefinitely

	var waitTime = 1

	for {
		select {

		case <-done:
			return connection

		default:
			log.L.Debugf("Creating salt connection")
			connection, err := login()

			if err == nil {
				return connection
			}

			log.L.Debugf("Error connecting to salt %v", err.Error())

			time.Sleep(time.Duration(waitTime) * time.Second)

			switch waitTime {
			case 1:
				waitTime = 10
			case 10:
				waitTime = 30
			case 30:
				waitTime = 60
			case 60:
				waitTime = 120
			case 120:
				waitTime = 300
			default:
			}
		}
	}
}

func readAndWriteEvent(line string, events chan string) *nerr.E {

	if strings.Contains(line, "data") {

		//cut off the "data:" so we just get the raw { } json value of the data field
		jsonString := line[5:]

		//We're not going to unmarshal here - instead we'll allow the translator
		//to unpack the json string into the appropriate struct based on the type
		//we're also not going to filter here, but instead allow the translator to do that
		//for unknown even types
		events <- jsonString
	}

	return nil
}
