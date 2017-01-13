package main

import (
	"flag"
	"fmt"
	"net/http"

	pxy "github.com/larryboymi/go-ocelot/proxy"
)

// Config type
type Config struct {
	serverPort      string
	serverPortUsage string
	targetURL       string
	targetUsage     string
}

func main() {
	config := Config{":8080", "default server port, ':8080', ':8081'...", "http://127.0.0.1:8081", "default redirect url, 'http://127.0.0.1:8080'"}

	port := flag.String("port", config.serverPort, config.serverPortUsage)
	url := flag.String("url", config.targetURL, config.targetUsage)

	flag.Parse()

	fmt.Println(fmt.Sprintf("running on : %s", *port))
	fmt.Println(fmt.Sprintf("redirect to : %s", *url))

	proxy := pxy.New(*url)

	http.HandleFunc("/", proxy.Handler)
	http.ListenAndServe(*port, nil)
}
