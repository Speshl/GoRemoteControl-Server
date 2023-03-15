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

	"github.com/Speshl/GoRemoteControl_Server/models"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	address           string
	serialPort        *string
	baudRate          *int
	videoDevice       *string
	useVideo          *bool
	videoPort         *string
	latestFrame       LatestFrame
	latestState       LatestState
	stopChannel       chan struct{} //stop video capture
	startChannel      chan struct{} //start video capture
	connectChannel    chan struct{} //stream request started
	disconnectChannel chan struct{} //stream request stopped

	fps uint32
	//pixfmt     v4l2.FourCCType
	width      int
	height     int
	streamInfo string
}

type PageData struct {
	StreamInfo  string
	StreamPath  string
	ImgWidth    int
	ImgHeight   int
	ControlPath string
}

func NewServer(address string, serialPort *string, baudRate *int, useVideo *bool, videoDevice *string, videoPort *string) *Server {

	return &Server{
		address:     address,
		serialPort:  serialPort,
		baudRate:    baudRate,
		useVideo:    useVideo,
		videoDevice: videoDevice,
		videoPort:   videoPort,

		fps:    30,
		width:  640,
		height: 480,
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	log.Println("starting controller server...")
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

	go s.startVideoServer(ctx)

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) startUDPListener(ctx context.Context) error {
	log.Printf("starting UDP listener on port %s...", s.address)
	defer log.Println("stopping UDP listener")
	addr := net.UDPAddr{
		Port: 1053,
		IP:   net.ParseIP("0.0.0.0"),
	}
	udpServer, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}

	defer func() {
		udpServer.Close()
		log.Println("UDP listener closing")
	}()

	log.Printf("listening on address %s...\n", s.address)
	ticker := time.NewTicker(5 * time.Second)
	var lastRecieved time.Time
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Printf("ticker log - latest UDP packet received: %s\n", lastRecieved.Format("2006-01-02 15:04:05"))
		default:
			buffer := make([]byte, 512)
			udpServer.SetDeadline(time.Now().Add(time.Second))
			numRead, _, err := udpServer.ReadFromUDP(buffer)
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
				//log.Printf("%d bytes (Type: %s) read from %s with delay %s (Sent: %d) (Now: %d)\n", numRead, packet.StateType, addr.String(), time.Since(packet.SentAt).String(), packet.SentAt, time.Now())
				s.latestState.Set(packet.State)
				lastRecieved = time.Now()
			}
		}
	}
}
