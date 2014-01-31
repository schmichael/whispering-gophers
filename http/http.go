package http

import (
	"encoding/json"
	"net/http"
)

var (
	Mux    = http.NewServeMux()
	Server = &http.Server{Handler: Mux}
)

func Serve(addr string, listFunc func() []string) {
	Server.Addr = addr
	Mux.HandleFunc("/peers", func(rw http.ResponseWriter, resp *http.Request) {
		out := &struct {
			Peers []string `json:"peers"`
		}{listFunc()}
		if buf, err := json.Marshal(out); err == nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.Write(buf)
		} else {
			rw.Header().Set("Content-Type", "text/plain")
			rw.Write([]byte(err.Error()))
		}
	})
	Server.ListenAndServe()
}
