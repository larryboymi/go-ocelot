package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	service "github.com/larryboymi/go-ocelot/api"
	"github.com/larryboymi/go-ocelot/middleware"
	"github.com/larryboymi/go-ocelot/proxy"
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

	proxy := proxy.New(&synchronizer)

	api := service.New(&synchronizer)

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Mux())
	mux.HandleFunc("/", proxy.Handler)

	loggedHandler := middleware.LoggedHandler(mux)
	headeredHandler := middleware.HeaderedHandler(loggedHandler)

	//  Start HTTP
	go func() {
		errHTTP := http.ListenAndServe(config.serverPort, headeredHandler)
		if errHTTP != nil {
			log.Fatal("HTTP Serving Error: ", errHTTP)
		}
	}()

	// Start TLS
	errTLS := http.ListenAndServeTLS(config.serverTLSPort, "cert.pem", "key.pem", headeredHandler)
	if errTLS != nil {
		log.Fatal("TLS Serving Error: ", errTLS)
	}
}
