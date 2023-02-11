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

	listJoysticks := flag.Bool("joys", false, "List available joysticks")
	listSerial := flag.Bool("serial", false, "List available serial devices")
	isServer := flag.Bool("server", true, "Run as UDP Server")
	isClient := flag.Bool("client", true, "Run as UDP Client and Controller Reader")
	udpPort := flag.String("port", "1053", "UDP Port")
	serialPort := flag.String("device", "", "Serial device")
	deviceCfg := flag.String("cfg", "./configs/g27.json", "Path to cfg json")

	if listJoysticks != nil && *listJoysticks {
		_, err := client.GetJoysticks()
		if err != nil {
			log.Fatal(err)
		}
	} else if listSerial != nil && *listSerial {
		err := server.GetSerialDevices()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		errorGroup, _ := errgroup.WithContext(context.Background())
		if isClient != nil && *isClient {
			c := client.NewClient(":"+*udpPort, *deviceCfg)
			errorGroup.Go(c.RunClient)
		}
		if isServer != nil && *isServer {
			s := server.NewServer(":"+*udpPort, serialPort)
			errorGroup.Go(s.RunServer)
		}
		err := errorGroup.Wait()
		if err != nil {
			log.Fatalf("Errorgroup had error: %s", err.Error())
		}
	}
}
