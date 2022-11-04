package main

import (
	"machine"
	"time"

	"github.com/soypat/achicken"
)

var (
	USB = machine.USBCDC
	LED = machine.LED
)

func main() {
	LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	var cmd achicken.CmdSerial
	cycle := true
	i := uint16(0)
	for {
		LED.Set(cycle)
		cycle = !cycle
		err := cmd.ReadNext(USB)
		if err != nil {
			cmd = achicken.NewCommand(2*i, 2*i+1)
			i++
			_, err = cmd.WriteTo(USB)
			if err != nil {
				println(err)
			}
		} else {
			verb, noun := cmd.Command()
			print("got command: ")
			println(verb, noun)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
