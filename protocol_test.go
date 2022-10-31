package achicken

import (
	"os"
	"testing"
	"time"
)

const deviceFile = "/dev/ttyACM0"

// Writes command to picolia if connected
func TestCommand(t *testing.T) {
	const bufSize = 400
	dev, err := os.Open(deviceFile)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	cmd := NewCommand(0x2e, 0xf2)
	msg := append([]byte{'{'}, cmd[:]...)
	t.Error(msg)
	go func() {
		for i := 0; i < 5; i++ {
			dev.Write(msg)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	bigMessage := make([]byte, bufSize)
	total := 0
	for total < bufSize {
		n, err := dev.Read(bigMessage[total:])
		if err != nil {
			t.Fatal(err)
		}
		total += n
	}
	t.Error(string(bigMessage[:total]))
}
