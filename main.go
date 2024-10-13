package main

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/iamNilotpal/drp/proxy"
	"github.com/iamNilotpal/drp/server"
)

func main() {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHostFromEnv(),
		client.WithVersionFromEnv(),
	)
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	appRouter := server.CreateRouter(cli)
	server := http.Server{Handler: appRouter, Addr: "localhost:8000"}

	proxyRouter := proxy.CreateReverseProxy()
	proxyServer := http.Server{Handler: proxyRouter, Addr: "localhost:80"}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Error starting server : %+v", err)
		}
	}()

	if err := proxyServer.ListenAndServe(); err != nil {
		fmt.Printf("Error starting proxy server : %+v", err)
	}
}
