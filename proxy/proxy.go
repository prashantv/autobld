package proxy

import (
	"fmt"
	"net"
	"time"

	"github.com/prashantv/autobld/log"
	"github.com/prashantv/autobld/sync"
)

func Start(config Config, errC chan<- error) {
	h := &tcpProxy{
		config: config,
		errC:   errC,
	}

	switch config.Type {
	case TCP:
		h.StartListen()
	case HTTP:
		hp := &httpProxy{tcpProxy: *h}
		hp.StartListen()
	default:
		errC <- fmt.Errorf("unknown proxy type: %v", config.Type)
	}
}

var TryConnect = sync.NewBool(true)

func (h *tcpProxy) connectPort(withRetry bool) (net.Conn, error) {
	const retryInterval = time.Millisecond * 200
	const maxRetries = 300
	numRetries := 10
	if withRetry {
		numRetries = maxRetries
	}
	addr := fmt.Sprintf("localhost:%v", h.config.ForwardTo)
	var errors []error
	for i := 0; i < numRetries; i++ {
		conn, err := net.DialTimeout("tcp", addr, retryInterval)
		if err == nil {
			// connection is valid
			return conn, nil
		}
		errors = append(errors, err)
		time.Sleep(retryInterval)
	}
	log.Log("connectWithRetry failed, first error: %v", errors[0:1])
	return nil, errors[len(errors)-1]
}
