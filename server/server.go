package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Speshl/GoRemoteControl/models"
	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	address    string
	serialPort *string
}

func NewServer(address string, serialPort *string) *Server {
	return &Server{
		address:    address,
		serialPort: serialPort,
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	log.Println("Starting Controller Server...")

	errGroup, ctx := errgroup.WithContext(ctx)
	stateChannel := s.startUDPListener(ctx, errGroup)
	latestState := s.startStateSyncer(ctx, errGroup, stateChannel)
	s.startSerialWriter(ctx, errGroup, latestState)

	err := errGroup.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) startUDPListener(ctx context.Context, errGroup *errgroup.Group) chan models.StateIface {
	returnChannel := make(chan models.StateIface, 5)
	errGroup.Go(func() error {
		udpServer, err := net.ListenPacket("udp", s.address)
		if err != nil {
			return err
		}

		defer func() {
			udpServer.Close()
			close(returnChannel)
			log.Println("UDP Listener closing")
		}()

		log.Println("Listening...")
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				buffer := make([]byte, 512)
				udpServer.SetDeadline(time.Now().Add(time.Second))
				numRead, _, err := udpServer.ReadFrom(buffer)
				if err != nil {
					if !errors.Is(err, os.ErrDeadlineExceeded) {
						log.Printf("server read error: %s\n", err.Error())
					}
					continue
				}
				if numRead > 0 {
					var packet models.Packet
					dec := gob.NewDecoder(bytes.NewReader(buffer))
					gob.Register(models.GroundState{})
					err := dec.Decode(&packet)
					if err != nil {
						log.Printf("server decode error: %s\n", err.Error())
						continue
					}
					//log.Printf("%d bytes (Type: %s) read from %s with delay %s\n", numRead, packet.StateType, addr.String(), time.Since(packet.SentAt).String())
					returnChannel <- packet.State
				}
			}
		}
	})
	return returnChannel
}

func (s *Server) startStateSyncer(ctx context.Context, errGroup *errgroup.Group, dataChannel <-chan models.StateIface) *LatestState {
	returnMutex := LatestState{}
	errGroup.Go(func() error {
		defer log.Println("State Syncer Closing")
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case data, ok := <-dataChannel:
				if !ok {
					return nil
				}
				returnMutex.Set(data)
			}
		}
	})
	return &returnMutex
}

func (s *Server) startSerialWriter(ctx context.Context, errGroup *errgroup.Group, latestState *LatestState) error {
	ticker := time.NewTicker(1000 * time.Millisecond) //RF Update rate
	errGroup.Go(func() error {
		defer log.Println("Serial Writer Closing")
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}
		for _, port := range ports {
			log.Printf("Found port: %v\n", port)
		}

		mode := &serial.Mode{
			BaudRate: 115200,
		}

		portName := ports[0]
		if s.serialPort != nil {
			portName = *s.serialPort
		}
		port, err := serial.Open(portName, mode)
		if err != nil {
			log.Fatal(err)
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				state, err := latestState.Get()
				if err != nil {
					log.Println("skipping rf send - latest state already used")
					continue
				}

				if state == nil {
					log.Println("got nil state")
					continue
				}

				stateBytes := state.GetBytes()
				log.Printf("State: %+v\n", stateBytes)
				_, err = port.Write(stateBytes)
				if err != nil {
					return err
				}
			}
		}
	})
	return nil
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
		log.Printf("Found port: %v\n", port)
	}
	return nil
}
