// Package fetcher implements a simple bridge for HTTP request and response
package fetcher

import (
	"appengine"
	"appengine/urlfetch"
	"bufio"
	"net/http"
)

func init() {
	http.HandleFunc("/", hello) // this will handle all except /http and /connect
	http.HandleFunc("/http", HandleHTTP)
	http.HandleFunc("/connect", HandleConnect)
}

// Flow:
//  1. read the real client request R1 from the body of request R2
//  2. round trip the request R1 by urlfetch
//  3. write response P1 of request R1 as the body of response P2
//  4. send back response P2
func HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// create context for gae request
	ctx := appengine.NewContext(r)
	cli := urlfetch.Client(ctx)
	s := NewSession(ctx, w)

	// read client requst
	creq, err := http.ReadRequest(bufio.NewReader(r.Body))
	if err != nil {
		s.HTTPError("ReadRequest: %s", err.Error())
		return
	}

	// record the real client request
	s.Info("%s %s %s\n", creq.Method, creq.URL.String(), creq.Proto)

	// round trip the client request
	// in fact RoundTrip supports both http and https
	resp, err := cli.Transport.RoundTrip(creq)
	if err != nil {
		s.HTTPError("RoundTrip: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// remote server response
	s.Info("RESPONSE %s\n", resp.Status)

	// write response and send to client
	if err := resp.Write(w); err != nil {
		s.HTTPError("Write: %s", err.Error())
		return
	}
}

// unimplemented
func HandleConnect(w http.ResponseWriter, r *http.Request) {
	// TODO: socket?
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
