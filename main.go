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

	listJoysticks := flag.Bool("listjoys", false, "List available joysticks")
	showJoyStats := flag.Bool("joystats", false, "Shows states of connected joysticks")

	udpPort := flag.String("joystickport", "1053", "Joystick Port")

	videoDevice := flag.String("videodevice", "/dev/video0", "Video Device (eg. /dev/video0)")
	videoPort := flag.String("videoport", "1054", "Video Port")
	useVideo := flag.Bool("video", true, "Start video capture: true or false")

	controlDeviceCfg := flag.String("cfg", "./configs/g27.json", "Path to cfg json")

	listSerial := flag.Bool("listserial", false, "List available serial devices")
	serialPort := flag.String("serial", "COM3", "Serial Port")
	baudRate := flag.Int("baudrate", 115200, "Serial baudrate")

	if listJoysticks != nil && *listJoysticks {
		_, err := client.GetJoysticks()
		if err != nil {
			log.Fatal(err)
		}
	} else if showJoyStats != nil && *showJoyStats {
		_, err := client.ShowJoyStats()
		if err != nil {
			log.Fatal(err)
		}
	} else if listSerial != nil && *listSerial {
		err := server.GetSerialDevices()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		errorGroup, ctx := errgroup.WithContext(context.Background())
		if isClient != nil && *isClient {
			c := client.NewClient(":"+*udpPort, *controlDeviceCfg)
			errorGroup.Go(func() error { return c.RunClient(ctx) })
		}
		if isServer != nil && *isServer {
			s := server.NewServer(":"+*udpPort, serialPort, baudRate, useVideo, videoDevice, videoPort)
			errorGroup.Go(func() error { return s.RunServer(ctx) })
		}
		err := errorGroup.Wait()
		if err != nil {
			log.Fatalf("Errorgroup had error: %s", err.Error())
		}
	}
}
