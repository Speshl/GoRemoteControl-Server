package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func (s *Server) startVideoCapture(ctx context.Context) error {
	if s.useVideo == nil || !*s.useVideo || s.videoDevice == nil || *s.videoDevice == "" {
		log.Println("skip starting video capture")
		return nil
	}

	log.Printf("starting camera %s...", *s.videoDevice)
	defer log.Printf("stopping camera")

	camera, err := device.Open(*s.videoDevice,
		device.WithPixFormat(v4l2.PixFormat{PixelFormat: v4l2.PixelFmtMJPEG, Width: 640, Height: 480}),
	)
	if err != nil {
		return err
	}
	defer camera.Close()

	if err := camera.Start(ctx); err != nil {
		return fmt.Errorf("camera start: %w", err)
	}

	caps := camera.Capability()
	log.Printf("device info: %s", caps.String())
	currFmt, err := camera.GetPixFormat()
	if err != nil {
		log.Fatalf("unable to get format: %s", err)
	}
	log.Printf("current camera format: %s", currFmt)

	s.streamInfo = fmt.Sprintf("%s - %s [%dx%d] %d fps",
		caps.Card,
		v4l2.PixelFormats[currFmt.PixelFormat],
		currFmt.Width, currFmt.Height, s.fps,
	)

	frames := camera.GetOutput()

	err = camera.Stop()
	if err != nil {
		return err
	}
	log.Printf("camera started successfully, now waiting for client")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChannel:
			err = camera.Stop()
			if err != nil {
				return err
			}
		case <-s.startChannel:
			err = camera.Start(ctx)
			if err != nil && err.Error() != "device: stream already started" {
				return err
			}
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
	log.Println("starting video server...")
	http.HandleFunc("/webcam", s.servePage) // returns an html page
	http.HandleFunc("/stream", s.streamVideo)
	log.Fatal(http.ListenAndServe(":"+*s.videoPort, nil))
	return nil
}

func (s *Server) streamVideo(w http.ResponseWriter, req *http.Request) {
	log.Println("got stream request")
	s.connectChannel <- struct{}{}

	mimeWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	ctx := req.Context()
	ticker := time.NewTicker(1 * time.Millisecond) //Video Update Rate
	for {
		select {
		case <-ctx.Done():
			s.disconnectChannel <- struct{}{}
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

func (s *Server) servePage(w http.ResponseWriter, r *http.Request) {
	pd := PageData{
		StreamInfo:  s.streamInfo,
		StreamPath:  fmt.Sprintf("/stream?%d", time.Now().UnixNano()),
		ImgWidth:    s.width,
		ImgHeight:   s.height,
		ControlPath: "/control",
	}

	// Start HTTP response
	w.Header().Add("Content-Type", "text/html")
	t, err := template.ParseFiles("viewer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("error serving page: %s", err.Error())
		return
	}

	// execute and return the template
	w.WriteHeader(http.StatusOK)
	err = t.Execute(w, pd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) startClientCounter(ctx context.Context) {
	numClients := 0
	cameraStarted := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.connectChannel:
			numClients++
		case <-s.disconnectChannel:
			numClients--
		}

		if cameraStarted && numClients <= 0 {
			s.stopChannel <- struct{}{} //turn off camera
			cameraStarted = !cameraStarted
		} else if !cameraStarted && numClients > 0 {
			s.startChannel <- struct{}{} //turn on camera
		}
	}
}
