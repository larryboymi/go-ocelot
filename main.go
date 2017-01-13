package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	pxy "github.com/larryboymi/go-ocelot/proxy"
)

// Config type
type Config struct {
	serverPort         string
	serverPortUsage    string
	serverTLSPort      string
	serverTLSPortUsage string
	targetURL          string
	targetUsage        string
}

func main() {
	config := &Config{
		serverPort:         ":8080",
		serverPortUsage:    "server port, ':8080'",
		serverTLSPort:      ":8443",
		serverTLSPortUsage: "server TLS port, ':8443'",
		targetURL:          "http://127.0.0.1:8081",
		targetUsage:        "redirect url, 'http://127.0.0.1:8081'",
	}

	port := flag.String("port", config.serverPort, config.serverPortUsage)
	tlsPort := flag.String("tlsPort", config.serverTLSPort, config.serverTLSPortUsage)
	url := flag.String("url", config.targetURL, config.targetUsage)

	flag.Parse()

	fmt.Println(fmt.Sprintf("running on HTTP: %s, TLS: %s", *port, *tlsPort))
	fmt.Println(fmt.Sprintf("redirect to : %s", *url))

	proxy := pxy.New(*url)

	http.HandleFunc("/", proxy.Handler)

	//  Start HTTP
	go func() {
		errHTTP := http.ListenAndServe(*port, nil)
		if errHTTP != nil {
			log.Fatal("HTTP Serving Error: ", errHTTP)
		}
	}()

	// Start TLS
	errTLS := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
	if errTLS != nil {
		log.Fatal("TLS Serving Error: ", errTLS)
	}
}
