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
	"strings"
	"time"

	"github.com/Speshl/GoRemoteControl/models"
	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	address    string
	serialPort *string
	baudRate   *int
}

func NewServer(address string, serialPort *string, baudRate *int) *Server {
	return &Server{
		address:    address,
		serialPort: serialPort,
		baudRate:   baudRate,
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	log.Println("Starting Controller Server...")

	errGroup, ctx := errgroup.WithContext(ctx)
	stateChannel := s.startUDPListener(ctx, errGroup)
	latestState := s.startStateSyncer(ctx, errGroup, stateChannel)
	s.startSerial(ctx, errGroup, latestState)

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("server error: %w", err)
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

func (s *Server) startSerial(ctx context.Context, errGroup *errgroup.Group, latestState *LatestState) error {
	errGroup.Go(func() error {
		defer log.Println("Serial Reader and Writer Started")
		serialPort, err := openSerialPort(s.serialPort, s.baudRate)
		if err != nil {
			return err
		}
		s.startSerialWriter(ctx, errGroup, &serialPort, latestState)
		s.startSerialReader(ctx, errGroup, &serialPort)
		return nil
	})
	return nil
}

func (s *Server) startSerialWriter(ctx context.Context, errGroup *errgroup.Group, serialPort *serial.Port, latestState *LatestState) error {
	ticker := time.NewTicker(5 * time.Millisecond) //RF Update rate
	errGroup.Go(func() error {
		defer log.Println("Serial Writer Closing")

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
				numSent, err := (*serialPort).Write(stateBytes)
				if err != nil {
					return fmt.Errorf("serial write error: %w", err)
				}
				if numSent != len(stateBytes) {
					log.Println("serial wrote wrong byte count")
				}
			}
		}
	})
	return nil
}

func (s *Server) startSerialReader(ctx context.Context, errGroup *errgroup.Group, serialPort *serial.Port) error {
	errGroup.Go(func() error {
		defer log.Println("Serial Reader Closing")
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
				log.Printf("Serial RX (%d bytes): %s", numRead, strings.TrimSpace(string(readBytes)))
			}
		}
	})
	return nil
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
		log.Printf("Found port: %v\n", port)
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
		log.Printf("Found port: %v\n", port)
	}
	return nil
}
