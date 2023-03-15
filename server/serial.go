package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

func (s *Server) startSerial(ctx context.Context) error {
	defer log.Println("Starting Serial Workers...")

	serialPort, err := openSerialPort(s.serialPort, s.baudRate)
	if err != nil {
		return err
	}

	log.Printf("starting serial workers on %s", *s.serialPort)

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return s.startSerialWriter(ctx, &serialPort)
	})

	errGroup.Go(func() error {
		return s.startSerialReader(ctx, &serialPort)
	})

	err = errGroup.Wait()
	if err != nil {
		return fmt.Errorf("serial error: %w", err)
	}

	return nil
}

func (s *Server) startSerialWriter(ctx context.Context, serialPort *serial.Port) error {
	ticker := time.NewTicker(30 * time.Millisecond) //RF Update rate
	log.Println("serial writer started")
	defer log.Println("serial writer closing")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			state, err := s.latestState.Get()
			if err != nil {
				//log.Println("skipping RF send - latest state already used")
				continue
			}

			if state == nil {
				log.Println("got nil control state")
				continue
			}

			triggerKey := []byte{255, 127} //prepended to data to keep in sync
			stateBytes := append(triggerKey, state.GetBytes()...)
			_, err = (*serialPort).Write(stateBytes)
			if err != nil {
				return fmt.Errorf("serial write error: %w", err)
			}
		}
	}
}

func (s *Server) startSerialReader(ctx context.Context, serialPort *serial.Port) error {
	log.Println("starting serial reader...")
	defer log.Println("serial reader closing")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			readBytes := make([]byte, 8096)
			numRead, err := (*serialPort).Read(readBytes)
			if err != nil {
				return fmt.Errorf("serial read error: %w", err)
			}
			log.Printf("serial RX (%d bytes): %s", numRead, strings.TrimSpace(string(readBytes)))
		}
	}
}

func openSerialPort(portParam *string, baudParam *int) (serial.Port, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no serial ports found")
	}
	for _, port := range ports {
		log.Printf("found port: %v\n", port)
	}

	baudRate := 115200
	if baudParam != nil {
		baudRate = *baudParam
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
	}

	portName := ports[0]
	paramFound := false
	if portParam != nil {
		for _, port := range ports {
			if port == *portParam {
				portName = port
				paramFound = true
			}
		}
		if !paramFound {
			return nil, fmt.Errorf("specified serial port not found: %s", *portParam)
		}
	}
	return serial.Open(portName, mode)
}

func GetSerialDevices() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		return fmt.Errorf("no serial ports found!")
	}
	for _, port := range ports {
		log.Printf("found port: %v\n", port)
	}
	return nil
}
