package proxy

import (
	"fmt"
	"io"
	"net"
)

type tcpProxy struct {
	proxy
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
	conn, err := h.connectPort(h.tryConnect.Read())
	if err != nil {
		h.errC <- err
		return
	}

	// Copy data between the connections, and block till at least one of them returns.
	done := make(chan struct{}, 1)
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
