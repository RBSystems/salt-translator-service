package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/salt-translator-service/salt"
)

func main() {

	const NUM_PROCESSES = 2

	events := make(chan salt.Event)
	done := make(chan bool, 1)
	signals := make(chan os.Signal)
	var control sync.WaitGroup

	signal.Notify(signals, os.Interrupt)
	go func() {
		log.Printf("Wating for interrupt")
		<-signals
		log.Printf("Nuclear launch detected. Firing interceptors...")
		for i := 0; i < NUM_PROCESSES; i++ {
			done <- true
		}
	}()

	go salt.Listen(events, done, &control)
	go elk.Publish(events, done, &control)

	control.Add(NUM_PROCESSES)
	control.Wait()
	log.Printf("Exiting...")
}
