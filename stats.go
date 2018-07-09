package stats

import (
	"fmt"
	"log"
	"net"
)

// conn for sending UDP packets
var sender *Sender

func init() {
	var err error
	sender, err = New("", "")
	if err != nil {
		log.Println("Trouble initializing default sender", err)
	}
}

// Count increments a stat.
func Count(name string, amount int) {
	if sender != nil {
		sender.Count(name, amount)
	}
}

// Value sets the value for a stat.
func Value(name string, value float64) {
	if sender != nil {
		sender.Value(name, value)
	}
}

// Sender creates UDP packets for each statistic update and transports them to a local statsd server.
type Sender struct {
	conn     net.Conn    // conn is the UDP connection for sending packets
	prefix   string      // prefix to add to all statistics names
	prefixed bool        // prefixed is true if there is a prefix
	suffix   string      // suffix to add to all statistics names
	suffixed bool        // suffixed is true if there is a suffix
	buf      chan string // buf buffers outgoing stats
}

// New creates a new Sender.
func New(prefix, suffix string) (*Sender, error) {
	return NewSender(prefix, suffix, ":9045", make(chan string, 1000))
}

// NewSender creates a new Sender with specific settings
func NewSender(prefix, suffix, target string, buffer chan string) (*Sender, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}

	s := &Sender{
		conn: conn,
	}
	s.Buffer(buffer)
	return s, nil
}

// Buffer configures the internal buffer to use a provided channel. If the channel
// is nil, the method returns the current buffer channel.
func (s *Sender) Buffer(buf chan string) chan string {
	if buf != nil {
		if s.buf != nil {
			close(s.buf) // Close the current buffer
		}
		s.buf = buf
		go s.start() // Start a go routine to process the buffer
	}
	return s.buf
}

// Count sends a counted statistic.
func (s *Sender) Count(name string, amount int) {
	// TODO benchmark and determine most efficient way to format the strings...
	if s.prefixed {
		if s.suffixed {
			s.buf <- fmt.Sprintf("%d=%s.%s.%s", amount, s.prefix, name, s.suffix)
		} else {
			s.buf <- fmt.Sprintf("%d=%s.%s", amount, s.prefix, name)
		}
	} else {
		if s.suffixed {
			s.buf <- fmt.Sprintf("%d=%s.%s", amount, name, s.suffix)
		} else {
			s.buf <- fmt.Sprintf("%d=%s", amount, name)
		}
	}
}

// Value sends a statistic with a set value.
func (s *Sender) Value(name string, amount float64) {
	if s.prefixed {
		if s.suffixed {
			s.buf <- fmt.Sprintf("#%f=%s.%s.%s", amount, s.prefix, name, s.suffix)
		} else {
			s.buf <- fmt.Sprintf("#%f=%s.%s", amount, s.prefix, name)
		}
	} else {
		if s.suffixed {
			s.buf <- fmt.Sprintf("#%f=%s.%s", amount, name, s.suffix)
		} else {
			s.buf <- fmt.Sprintf("#%f=%s", amount, name)
		}
	}
}

// start the Sender proessing the current buffer (blocks until the buffer is closed).
func (s *Sender) start() error {
	buf := s.buf
	for {
		stat, ok := <-buf
		if ok {
			_, err := s.conn.Write([]byte(stat))
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}
