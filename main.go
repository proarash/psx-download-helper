package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

const targetHost = "gst.prod.dl.playstation.net"

func isTarget(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	return strings.EqualFold(host, targetHost)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHTTPS(w, r)
		return
	}

	handleHTTP(w, r)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	if isTarget(r.Host) {
		log.Println(r.Method, r.URL.String())
	}

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	if isTarget(r.Host) {
		log.Println("CONNECT", r.Host)
	}

	dest, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer dest.Close()

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}

	client, _, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer client.Close()

	go io.Copy(dest, client)
	io.Copy(client, dest)
}

func main() {
	log.Println("Proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handle)))
}
