package proxy

import (
	"fmt"
	"io"
	"net"
)

type tcpProxy struct {
	config Config
	errC   chan<- error
}

func (h *tcpProxy) StartListen() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", h.config.Port))
	if err != nil {
		h.errC <- err
		return
	}

	h.Listen(listener)
}

func (h *tcpProxy) Listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			h.errC <- err
			return
		}
		go h.Handle(conn)
	}
}
func (h *tcpProxy) Handle(l net.Conn) {
	retry := TryConnect.Read()
	conn, err := h.connectPort(retry)
	if err != nil {
		h.errC <- err
		return
	}
	if retry {
		TryConnect.Write(false)
	}

	// Copy the data between the conn
	done := make(chan struct{})
	go func() {
		io.Copy(l, conn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(conn, l)
		done <- struct{}{}
	}()

	<-done
	l.Close()
	conn.Close()
}
