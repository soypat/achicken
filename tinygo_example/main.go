package main

import (
	"machine"
	"time"

	"github.com/soypat/achicken"
)

var USB = machine.USBCDC

func main() {
	var cmd achicken.CmdSerial
	for {
		err := cmd.ReadNext(USB)
		if err != nil {
			println(err)
		} else {
			verb, noun := cmd.Command()
			print("got command:")
			println(verb)
			println(noun)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
