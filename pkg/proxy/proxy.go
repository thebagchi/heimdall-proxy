package proxy

import (
	"net"
)

func Loop(local, forward, duplicate *Connection, done chan<- bool) {
	if !local.Valid() {
		if local.mandatory {
			done <- true
		}
		return
	}
	if local.Valid() {
		if local.mandatory {
			defer func() {
				done <- true
			}()
		}
	}
	buffer := make([]byte, 1500)
	for {
		read, err := local.Reader().Read(buffer)
		if nil != err {
			local.Deactivate()
			break
		}
		connections := []*Connection{
			forward, duplicate,
		}
		erred := false
		if read > 0 {
			data := buffer[:read]
			for i := 0; i < 2; i++ {
				conn := connections[i]
				if nil == conn {
					erred = true
					break
				}
				if !conn.IsActive() && conn.Mandatory() {
					erred = true
					break
				}
				if _, err := conn.Writer().Write(data); nil != err {
					if conn.Mandatory() {
						erred = true
						break
					}
				} else {
					_ = conn.Writer().Flush()
				}
			}
		}
		if erred {
			break
		}
	}
}

func Handle(conn net.Conn, where, also string) {
	defer func() {
		recover()
	}()
	connection := NewConnection(conn, true)
	defer connection.Close()

	var forwarder *Connection = nil
	{
		conn, err := net.Dial("tcp", where)
		if nil != err {
			return
		}
		forwarder = NewConnection(conn, true)
		if nil != err {
			_ = conn.Close()
			return
		}
	}
	defer forwarder.Close()

	var duplicator *Connection = nil
	{
		conn, err := net.Dial("tcp", also)
		if nil != err {
			duplicator = NewConnection(nil, false)
		} else {
			duplicator = NewConnection(conn, false)
		}
	}
	defer duplicator.Close()

	done := make(chan bool, 1)

	go Loop(connection, forwarder, duplicator, done)
	go Loop(forwarder, connection, nil, done)
	go Loop(duplicator, nil, nil, done)

	<-done
}

func StartServer(from, where, also string) error {
	connection, err := net.Listen("tcp", from)
	if nil != err {
		return err
	}
	defer connection.Close()
	for {
		local, err := connection.Accept()
		if err != nil {
			continue
		}
		go Handle(local, where, also)
	}
}
