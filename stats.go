package stats

import (
	"fmt"
	"net"

	"github.com/apex/log"
)

// conn for sending UDP packets
var conn net.Conn

func init() {
	var err error
	conn, err = net.Dial("udp", ":9045")
	if err != nil {
		log.WithError(err).Info("dial")
	}
}

// Count increments a stat.
func Count(name string, amount int) {
	if conn != nil {
		conn.Write([]byte(fmt.Sprintf("%d=%s", amount, name)))
	}
}

// Value sets the value for a stat.
func Value(name string, value float64) {
	if conn != nil {
		conn.Write([]byte(fmt.Sprintf("#%f=%s", value, name)))
	}
}
