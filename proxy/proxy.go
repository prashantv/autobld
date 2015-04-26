package proxy

import (
	"fmt"
	"net"
	"time"

	"github.com/prashantv/autobld/log"
	"github.com/prashantv/autobld/sync"
)

type proxy struct {
	config     Config
	errC       chan<- error
	tryConnect *sync.Bool
}

var proxies []*proxy

// Start creates a goroutine for the given proxy Config.
func Start(config Config, errC chan<- error) {
	p := &proxy{
		config:     config,
		errC:       errC,
		tryConnect: sync.NewBool(true),
	}
	proxies = append(proxies, p)

	switch config.Type {
	case TCP:
		tp := &tcpProxy{*p}
		go tp.StartListen()
	case HTTP:
		hp := &httpProxy{proxy: *p}
		go hp.StartListen()
	default:
		errC <- fmt.Errorf("unknown proxy type: %v", config.Type)
	}
}

// RetryConnect sets tryConnect on all the proxies.
func RetryConnect() {
	for _, p := range proxies {
		p.tryConnect.Write(true)
	}
}

func (p *proxy) connectPort(withRetry bool) (net.Conn, error) {
	const retryInterval = time.Millisecond * 200
	const maxRetries = 300
	numRetries := 10
	if withRetry {
		numRetries = maxRetries
	}
	addr := fmt.Sprintf("localhost:%v", p.config.ForwardTo)
	var errors []error
	for i := 0; i < numRetries; i++ {
		conn, err := net.DialTimeout("tcp", addr, retryInterval)
		if err == nil {
			// connection is valid, update tryConnect
			if withRetry {
				p.tryConnect.Write(false)
			}
			return conn, nil
		}
		errors = append(errors, err)
		time.Sleep(retryInterval)
	}
	log.L("connectWithRetry failed, first error: %v", errors[0:1])
	return nil, errors[len(errors)-1]
}
