package main

import (
	"sync"

	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/salt-translator-service/salt"
)

func main() {

	events := make(chan salt.Event)
	done := make(chan bool, 1)
	var control sync.WaitGroup

	//Listen to salt
	salt.Listen(events, done)
	control.Add(1)

	//Publish to ELK
	elk.Publish(events, done)
	control.Add(1)

}
