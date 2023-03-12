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
	"golang.org/x/sync/errgroup"
)

type Server struct {
	address     string
	serialPort  *string
	baudRate    *int
	videoDevice *string
	useVideo    *bool
	videoPort   *string
	latestFrame LatestFrame
	latestState LatestState
}

func NewServer(address string, serialPort *string, baudRate *int, useVideo *bool, videoDevice *string, videoPort *string) *Server {
	return &Server{
		address:     address,
		serialPort:  serialPort,
		baudRate:    baudRate,
		useVideo:    useVideo,
		videoDevice: videoDevice,
		videoPort:   videoPort,
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	log.Println("starint controller server...")
	defer log.Println("controller server stopped")

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return s.startUDPListener(ctx)
	})

	errGroup.Go(func() error {
		return s.startSerial(ctx)
	})

	errGroup.Go(func() error {
		return s.startVideoCapture(ctx)
	})

	errGroup.Go(func() error {
		return s.startVideoServer(ctx)
	})

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) startUDPListener(ctx context.Context) error {
	log.Println("starting UDP listener...")
	udpServer, err := net.ListenPacket("udp", s.address)
	if err != nil {
		return err
	}

	defer func() {
		udpServer.Close()
		log.Println("UDP listener closing")
	}()

	log.Printf("listening on address %s...\n", s.address)
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
				s.latestState.Set(packet.State)
			}
		}
	}
}
