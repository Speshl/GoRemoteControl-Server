package main

import (
	"context"
	"flag"
	"log"

	"github.com/Speshl/GoRemoteControl_Server/server"
	"golang.org/x/sync/errgroup"
)

func main() {
	udpPort := flag.String("joystickport", "1053", "Joystick Port")

	videoDevice := flag.String("videodevice", "/dev/video0", "Video Device (eg. /dev/video0)")
	videoPort := flag.String("videoport", "1054", "Video Port")
	useVideo := flag.Bool("video", true, "Start video capture: true or false")

	listSerial := flag.Bool("listserial", false, "List available serial devices")
	serialPort := flag.String("serial", "COM3", "Serial Port")
	baudRate := flag.Int("baudrate", 115200, "Serial baudrate")

	flag.Parse()

	if listSerial != nil && *listSerial {
		err := server.GetSerialDevices()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		errorGroup, ctx := errgroup.WithContext(context.Background())

		s := server.NewServer(":"+*udpPort, serialPort, baudRate, useVideo, videoDevice, videoPort)
		errorGroup.Go(func() error { return s.RunServer(ctx) })

		err := errorGroup.Wait()
		if err != nil {
			log.Fatalf("Errorgroup had error: %s", err.Error())
		}
	}
}
