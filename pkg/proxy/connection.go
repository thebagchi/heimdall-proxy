package proxy

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"
)

// Connection object
// active: 0/1 0 = active 1 = inactive
type Connection struct {
	net.Conn
	active    int32
	reader    *bufio.Reader
	writer    *bufio.Writer
	valid     bool
	mandatory bool
}

func (conn *Connection) IsActive() bool {
	return atomic.LoadInt32(&conn.active) == 0
}

func (conn *Connection) Deactivate() bool {
	return atomic.CompareAndSwapInt32(&conn.active, 0, 1)
}

func (conn *Connection) Connection() net.Conn {
	return conn.Conn
}

func (conn *Connection) Connected() bool {
	return nil != conn.Conn
}

func (conn *Connection) Valid() bool {
	return conn.valid
}

func (conn *Connection) Mandatory() bool {
	return conn.mandatory
}

func (conn *Connection) Reader() *bufio.Reader {
	return conn.reader
}

func (conn *Connection) Writer() *bufio.Writer {
	return conn.writer
}

// Wrap net.Conn to Connection object
func NewConnection(connection net.Conn, mandatory bool) *Connection {
	if nil == connection {
		return &Connection{
			Conn:      nil,
			active:    1,
			valid:     false,
			mandatory: mandatory,
		}
	}
	if conn, ok := connection.(*net.TCPConn); ok {
		_ = conn.SetKeepAlive(true)
		_ = conn.SetKeepAlivePeriod(30 * time.Second)
		_ = conn.SetNoDelay(true)
	}
	return &Connection{
		Conn:      connection,
		active:    0,
		reader:    bufio.NewReader(connection),
		writer:    bufio.NewWriter(connection),
		valid:     true,
		mandatory: mandatory,
	}

}
