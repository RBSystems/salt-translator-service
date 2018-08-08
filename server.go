package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/salt-translator-service/salt"
)

func main() {
	const numProcesses = 3

	//log.SetLevel("debug")

	events := make(chan string)
	done := make(chan bool, 1)
	signals := make(chan os.Signal)

	signal.Notify(signals, os.Interrupt)
	go func() {
		log.L.Debugf("Waiting for interrupt")
		<-signals
		log.L.Debugf("sending interrupt signal to go routines")
		for i := 0; i < numProcesses; i++ {
			done <- true
		}
	}()

	go salt.Listen(events, done)
	go elk.Publish(events, done)

	go startServer()

	<-done
}

func startServer() {
	router := common.NewRouter()

	server := http.Server{
		Addr:           ":6997",
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
