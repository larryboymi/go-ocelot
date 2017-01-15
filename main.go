package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/larryboymi/go-ocelot/routes"
)

// Config type
type Config struct {
	serverPort    string
	serverTLSPort string
}

func main() {
	start(os.Args)
}

func start(args []string) {
	config := &Config{
		serverPort:    "0.0.0.0:8080",
		serverTLSPort: "0.0.0.0:8443",
	}

	redisURL := flag.String("redisURL", "redis:6379", "redis url, 'redis:6379'")

	flag.Parse()

	fmt.Println(fmt.Sprintf("running on HTTP: %s, TLS: %s", config.serverPort, config.serverTLSPort))

	//  Start Route Synchronizer
	synchronizer := routes.New(10, *redisURL)
	synchronizer.Start()

	http.HandleFunc("/", synchronizer.Handler)

	//  Start HTTP
	go func() {
		errHTTP := http.ListenAndServe(config.serverPort, nil)
		if errHTTP != nil {
			log.Fatal("HTTP Serving Error: ", errHTTP)
		}
	}()

	// Start TLS
	errTLS := http.ListenAndServeTLS(config.serverTLSPort, "cert.pem", "key.pem", nil)
	if errTLS != nil {
		log.Fatal("TLS Serving Error: ", errTLS)
	}
}
