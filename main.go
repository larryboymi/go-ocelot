package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	pxy "github.com/larryboymi/go-ocelot/proxy"
	"github.com/larryboymi/go-ocelot/routes"
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
	start(os.Args)
}

func start(args []string) {
	config := &Config{
		serverPort:         ":8080",
		serverPortUsage:    "server port, ':8080'",
		serverTLSPort:      ":8443",
		serverTLSPortUsage: "server TLS port, ':8443'",
		targetURL:          "http://ecgo:8081",
		targetUsage:        "redirect url, 'http://ecgo:8081'",
	}

	redisURL := flag.String("redisURL", config.serverPort, config.serverPortUsage)

	flag.Parse()

	fmt.Println(fmt.Sprintf("running on HTTP: %s, TLS: %s", config.serverPort, config.serverTLSPort))
	fmt.Println(fmt.Sprintf("redirect to : %s", config.targetURL))

	proxy := pxy.New(config.targetURL)

	http.HandleFunc("/", proxy.Handler)

	//  Start HTTP
	go func() {
		errHTTP := http.ListenAndServe(config.serverPort, nil)
		if errHTTP != nil {
			log.Fatal("HTTP Serving Error: ", errHTTP)
		}
	}()

	//  Start Route Maintainer
	go func() {
		routeMaintainer := routes.New(10, *redisURL)
		routeMaintainer.Start()
	}()

	// Start TLS
	errTLS := http.ListenAndServeTLS(config.serverTLSPort, "cert.pem", "key.pem", nil)
	if errTLS != nil {
		log.Fatal("TLS Serving Error: ", errTLS)
	}
}
