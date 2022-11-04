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
	if err != nil {
		t.Fatal(err, &cmd)
	}
	i := 0
	actual, _ := cmd.Command()
	for err == nil {
		i++
		err = cmd.ReadNext(ser)
		if err != nil {
			t.Error(err)
		}
		expected := NewCommand(actual+2, actual+3)
		if expected != cmd {
			t.Errorf("#%d expected command (%s), got (%s)", i, &expected, &cmd)
		}
		actual += 2
	}
	t.Error(b.String())
}
