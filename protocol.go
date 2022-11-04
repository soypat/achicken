package achicken

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

// CmdSerial represents a 32bit serialized command.
type CmdSerial [6]byte

func NewCommand(verb, noun uint16) (cmd CmdSerial) {
	binary.LittleEndian.PutUint16(cmd[0:2], verb)
	binary.LittleEndian.PutUint16(cmd[2:4], noun)
	crc := cmd.CalculateCRC()
	cmd.SetCRC(crc)
	if cmd.CRC() != crc {
		panic("achicken fatal: CRC was not set/unmarshalled correctly")
	}
	return cmd
}

var (
	ErrCmdSerialNotFound = errors.New("no command found in buffer")
	ErrBadCRC            = errors.New("bad crc")
	ErrBufferEmpty       = errors.New("serial buffer empty")
)

type Serialer interface {
	Buffered() int
	ReadByte() (byte, error)
}

var initiator = []byte{'{'}

// WriteTo marshals CmdSerial onto the writer in binary format.
func (c *CmdSerial) WriteTo(w io.Writer) (int64, error) {
	_, err := w.Write(initiator)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(c[:])
	return int64(n + 1), err
}

func (c *CmdSerial) ReadNext(serialer Serialer) error {
	n := serialer.Buffered()
	if n == 0 {
		return ErrBufferEmpty
	}
	if n < len(CmdSerial{}) {
		return ErrCmdSerialNotFound // Not enough data to form a command. wait for more data.
	}
	serIdx := -1
	for i := 0; i < n; i++ {
		isLast := i == n-1
		b, err := serialer.ReadByte()
		switch {
		case err != nil && !(isLast && err == io.EOF):
			return err
		case serIdx == -1 && n-i < len(CmdSerial{}):
			// Not enough data in buffer to form a command, wait for more data.
			return ErrCmdSerialNotFound
		case b == '{' && serIdx == -1:
			// Command start found.
			serIdx = 0
		case serIdx >= 0:
			// Continue unserializing command.
			c[serIdx] = b
			serIdx++
			if serIdx < len(CmdSerial{}) {
				// Command not yet fully unserialized.
				continue
			}

			// Reset state.
			serIdx = -1
			if c.CalculateCRC() != c.CRC() {
				// Failed validation check. Discard command.
				return ErrBadCRC
			}
			return nil
		}
	}
	return ErrCmdSerialNotFound
}

func (c *CmdSerial) Command() (verb, noun uint16) {
	verb = binary.LittleEndian.Uint16(c[0:2])
	noun = binary.LittleEndian.Uint16(c[2:4])
	return verb, noun
}

func (c *CmdSerial) CalculateCRC() uint16 {
	crc := NewCRC()
	crc = crc.AddByte(c[0])
	crc = crc.AddByte(c[1])
	crc = crc.AddByte(c[2])
	crc = crc.AddByte(c[3])
	return uint16(crc)
}

func (c *CmdSerial) CRC() uint16 {
	return binary.LittleEndian.Uint16(c[4:6])
}

func (c *CmdSerial) SetCRC(crc uint16) {
	binary.LittleEndian.PutUint16(c[4:6], crc)
}

func (c *CmdSerial) String() string {
	v, n := c.Command()
	crc := c.CRC()
	if crc != c.CalculateCRC() {
		return "bad crc"
	}
	return strconv.FormatUint(uint64(v), 16) + " " + strconv.FormatUint(uint64(n), 16) + " " + strconv.FormatUint(uint64(crc), 16)
}

// CRC is a port of the OpenCyphal spec CRC.
type CRC uint16

func NewCRC() CRC { return CRC(0xffff) }

func (c CRC) Add(data []byte) CRC {
	for _, b := range data {
		c = c.AddByte(b)
	}
	return c
}

// AddByte adds b to the CRC and returns the result.
//
//go:inline
func (c CRC) AddByte(b byte) CRC {
	return (c << 8) ^ crcTable[byte(c>>8)^b]
}

var crcTable = [256]CRC{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50A5, 0x60C6, 0x70E7, 0x8108, 0x9129, 0xA14A, 0xB16B, 0xC18C,
	0xD1AD, 0xE1CE, 0xF1EF, 0x1231, 0x0210, 0x3273, 0x2252, 0x52B5, 0x4294, 0x72F7, 0x62D6, 0x9339, 0x8318,
	0xB37B, 0xA35A, 0xD3BD, 0xC39C, 0xF3FF, 0xE3DE, 0x2462, 0x3443, 0x0420, 0x1401, 0x64E6, 0x74C7, 0x44A4,
	0x5485, 0xA56A, 0xB54B, 0x8528, 0x9509, 0xE5EE, 0xF5CF, 0xC5AC, 0xD58D, 0x3653, 0x2672, 0x1611, 0x0630,
	0x76D7, 0x66F6, 0x5695, 0x46B4, 0xB75B, 0xA77A, 0x9719, 0x8738, 0xF7DF, 0xE7FE, 0xD79D, 0xC7BC, 0x48C4,
	0x58E5, 0x6886, 0x78A7, 0x0840, 0x1861, 0x2802, 0x3823, 0xC9CC, 0xD9ED, 0xE98E, 0xF9AF, 0x8948, 0x9969,
	0xA90A, 0xB92B, 0x5AF5, 0x4AD4, 0x7AB7, 0x6A96, 0x1A71, 0x0A50, 0x3A33, 0x2A12, 0xDBFD, 0xCBDC, 0xFBBF,
	0xEB9E, 0x9B79, 0x8B58, 0xBB3B, 0xAB1A, 0x6CA6, 0x7C87, 0x4CE4, 0x5CC5, 0x2C22, 0x3C03, 0x0C60, 0x1C41,
	0xEDAE, 0xFD8F, 0xCDEC, 0xDDCD, 0xAD2A, 0xBD0B, 0x8D68, 0x9D49, 0x7E97, 0x6EB6, 0x5ED5, 0x4EF4, 0x3E13,
	0x2E32, 0x1E51, 0x0E70, 0xFF9F, 0xEFBE, 0xDFDD, 0xCFFC, 0xBF1B, 0xAF3A, 0x9F59, 0x8F78, 0x9188, 0x81A9,
	0xB1CA, 0xA1EB, 0xD10C, 0xC12D, 0xF14E, 0xE16F, 0x1080, 0x00A1, 0x30C2, 0x20E3, 0x5004, 0x4025, 0x7046,
	0x6067, 0x83B9, 0x9398, 0xA3FB, 0xB3DA, 0xC33D, 0xD31C, 0xE37F, 0xF35E, 0x02B1, 0x1290, 0x22F3, 0x32D2,
	0x4235, 0x5214, 0x6277, 0x7256, 0xB5EA, 0xA5CB, 0x95A8, 0x8589, 0xF56E, 0xE54F, 0xD52C, 0xC50D, 0x34E2,
	0x24C3, 0x14A0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405, 0xA7DB, 0xB7FA, 0x8799, 0x97B8, 0xE75F, 0xF77E,
	0xC71D, 0xD73C, 0x26D3, 0x36F2, 0x0691, 0x16B0, 0x6657, 0x7676, 0x4615, 0x5634, 0xD94C, 0xC96D, 0xF90E,
	0xE92F, 0x99C8, 0x89E9, 0xB98A, 0xA9AB, 0x5844, 0x4865, 0x7806, 0x6827, 0x18C0, 0x08E1, 0x3882, 0x28A3,
	0xCB7D, 0xDB5C, 0xEB3F, 0xFB1E, 0x8BF9, 0x9BD8, 0xABBB, 0xBB9A, 0x4A75, 0x5A54, 0x6A37, 0x7A16, 0x0AF1,
	0x1AD0, 0x2AB3, 0x3A92, 0xFD2E, 0xED0F, 0xDD6C, 0xCD4D, 0xBDAA, 0xAD8B, 0x9DE8, 0x8DC9, 0x7C26, 0x6C07,
	0x5C64, 0x4C45, 0x3CA2, 0x2C83, 0x1CE0, 0x0CC1, 0xEF1F, 0xFF3E, 0xCF5D, 0xDF7C, 0xAF9B, 0xBFBA, 0x8FD9,
	0x9FF8, 0x6E17, 0x7E36, 0x4E55, 0x5E74, 0x2E93, 0x3EB2, 0x0ED1, 0x1EF0,
}

// ReaderFromSerialer is probably not a good idea. Use only for debugging.
func ReaderFromSerialer(s Serialer) io.Reader {
	return serialReader{s: s}
}

type serialReader struct {
	s Serialer
}

func (r serialReader) Read(b []byte) (int, error) {
	n := r.s.Buffered()
	if n == 0 {
		// Should io.EOF be returned in this case? Nil? ErrBufferEmpty? Dunno.
		return 0, io.EOF
	} else if n > len(b) {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		char, err := r.s.ReadByte()
		if err != nil {
			return i, err
		}
		b[i] = char
	}
	return n, nil
}

func SerializeBytes(buf []byte) Serialer {
	return &byteSerialer{buf: buf}
}

type byteSerialer struct {
	buf []byte
}

func (b *byteSerialer) Buffered() int {
	return len(b.buf)
}

func (b *byteSerialer) ReadByte() (byte, error) {
	if b.Buffered() == 0 {
		b.buf = nil
		return 0, ErrBufferEmpty
	}
	char := b.buf[0]
	b.buf = b.buf[1:]
	return char, nil
}

func SerialerFromReader(r io.Reader) Serialer {
	return &serialerReader{
		r: r,
	}
}

type serialerReader struct {
	r   io.Reader
	buf bytes.Buffer
}

func (r *serialerReader) ReadByte() (byte, error) {
	return r.buf.ReadByte()
}

func (r *serialerReader) Buffered() int {
	var inbuf [64]byte
	n := 1
	i := 0
	var err error
	for err == nil && n != 0 && i < 6 {
		n, err = r.r.Read(inbuf[:])
		r.buf.Write(inbuf[:n])
		i++
	}
	return r.buf.Len()
}
