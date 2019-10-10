package proxy

import (
	"bufio"
	"net"
	"sync/atomic"
	"time"
)

// A Connection represents object that holds a net.Conn
// active:    0/1 0 = active 1 = inactive
// reader:    read from net.Conn
// writer:    writer into net.Conn
// valid:     underlying connection is valid
// mandatory: connection is mandatory
type Connection struct {
	net.Conn
	active    int32
	reader    *bufio.Reader
	writer    *bufio.Writer
	valid     bool
	mandatory bool
}

func (conn *Connection) Close() {
	_ = conn.Conn.Close()
}

func (conn *Connection) Read(buffer []byte) (int, error) {
	count, err := conn.reader.Read(buffer)
	if nil != err {
		atomic.CompareAndSwapInt32(&conn.active, 0, 1)
	}
	return count, err
}

func (conn *Connection) Write(buffer []byte) (int, error) {
	count, err := conn.writer.Write(buffer)
	if nil != err {
		_ = conn.writer.Flush()
	}
	return count, err
}

func (conn *Connection) IsActive() bool {
	return atomic.LoadInt32(&conn.active) == 0
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
