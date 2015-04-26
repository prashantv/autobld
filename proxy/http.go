package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type httpProxy struct {
	proxy
	rp *httputil.ReverseProxy
}

func (h *httpProxy) StartListen() {
	url, _ := url.Parse(fmt.Sprintf("http://localhost:%v/%s", h.config.ForwardTo, h.config.HTTPPath))
	h.rp = httputil.NewSingleHostReverseProxy(url)
	http.ListenAndServe(fmt.Sprintf(":%v", h.config.Port), h)
}

func (h *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.blockRequests.Wait()
	if h.tryConnect.Read() {
		if _, err := h.connectPort(true /* withRetry */); err != nil {
			writeErr(w, err)
			return
		}
	}
	h.rp.ServeHTTP(w, r)
}

func writeErr(w http.ResponseWriter, err error) {
	w.Write([]byte(fmt.Sprintf("Error connecting to underlying server: %v", err)))
}
