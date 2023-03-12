package server

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	videoDevice "github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func (s *Server) startVideoCapture(ctx context.Context) error {
	if s.useVideo == nil || !*s.useVideo || s.videoDevice == nil || *s.videoDevice == "" {
		log.Println("skip starting video capture")
		return nil
	}

	log.Printf("starting camera %s...", *s.videoDevice)
	defer log.Printf("")

	camera, err := videoDevice.Open(*s.videoDevice,
		videoDevice.WithPixFormat(v4l2.PixFormat{PixelFormat: v4l2.PixelFmtMJPEG, Width: 640, Height: 480}),
	)
	if err != nil {
		return err
	}
	defer camera.Close()

	if err := camera.Start(ctx); err != nil {
		return fmt.Errorf("camera start: %w", err)
	}

	frames := camera.GetOutput()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case frame, ok := <-frames:
			if !ok { //channel closed
				return nil
			}
			s.latestFrame.Set(frame)
		}
	}
}

func (s *Server) startVideoServer(ctx context.Context) error {
	if s.useVideo == nil || !*s.useVideo || s.videoPort == nil || *s.videoPort == "" {
		log.Println("skip starting video server")
		return nil
	}
	http.HandleFunc("/stream", s.streamVideo)
	log.Fatal(http.ListenAndServe(*s.videoPort, nil))
}

func (s *Server) streamVideo(w http.ResponseWriter, req *http.Request) {
	mimeWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	ctx := req.Context()
	ticker := time.NewTicker(30 * time.Millisecond) //Video Update Rate
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			frame, err := s.latestFrame.Get()
			if err != nil {
				continue // frame already sent
			}

			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Printf("failed to create multi-part writer: %s", err)
				return
			}

			if _, err := partWriter.Write(frame); err != nil {
				log.Printf("failed to write image: %s", err)
			}
		}
	}
}
