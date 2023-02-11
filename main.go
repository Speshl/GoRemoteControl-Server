package main

import (
	"context"
	"flag"
	"log"

	"github.com/Speshl/GoRemoteControl/client"
	"github.com/Speshl/GoRemoteControl/server"
	"golang.org/x/sync/errgroup"
)

func main() {

	isServer := flag.Bool("server", true, "Run as UDP Server")
	isClient := flag.Bool("client", true, "Run as UDP Client and Controller Reader")
	udpPort := flag.String("port", "1053", "UDP Port")
	deviceCfg := flag.String("cfg", "./configs/g27.json", "Path to cfg json")

	errorGroup, _ := errgroup.WithContext(context.Background())
	if isClient != nil && *isClient {
		client := client.NewClient(":"+*udpPort, *deviceCfg)
		errorGroup.Go(client.RunClient)
	}

	if isServer != nil && *isServer {
		server := server.NewServer(":" + *udpPort)
		errorGroup.Go(server.RunServer)
	}
	err := errorGroup.Wait()
	if err != nil {
		log.Fatalf("Errorgroup had error: %s", err.Error())
	}
}
