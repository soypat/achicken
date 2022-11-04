package achicken

import (
	"bytes"
	"io"
	"testing"
	"time"

	"go.bug.st/serial"
)

const deviceFile = "/dev/ttyACM0"

func TestCountUp(t *testing.T) {
	const bufSize = 400
	port, err := serial.Open(deviceFile, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer port.Close()
	totalRead := 0
	ser := SerialerFromReader(port)
	var cmd CmdSerial
	for i := 0; i < 50 && totalRead < bufSize-1; i++ {
		err := cmd.ReadNext(ser)
		if err != nil {
			t.Log(err)
		} else {
			t.Log("Command received:", cmd)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLoopback(t *testing.T) {
	const bufSize = 400
	port, err := serial.Open(deviceFile, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer port.Close()
	totalRead := 0
	var b bytes.Buffer
	for i := 0; i < 100 && totalRead < bufSize-1; i++ {
		buf := make([]byte, 128)
		n, err := port.Read(buf)
		if err == nil || err == io.EOF {
			b.Write(buf[:n])
		}
	}
	var cmd CmdSerial
	ser := SerialerFromReader(&b)
	err = cmd.ReadNext(ser)
	actual, ap1 := cmd.Command()

	if err != nil {
		t.Fatal(actual, ap1, cmd.CRC(), cmd.CalculateCRC(), err)
	}

	for err != nil {
		err = cmd.ReadNext(ser)
		if err != nil {
			t.Error(err)
		}
		expect := actual + 2
		got, gotp1 := cmd.Command()
		if got != expect {
			t.Errorf("expected command #%d, got #%d", expect, got)
		}
		if gotp1 != got+1 {
			t.Errorf("expected noun %d to be verb+1, got %d", got+1, gotp1)
		}
		actual = got
	}
	t.Error(b.String())
}
