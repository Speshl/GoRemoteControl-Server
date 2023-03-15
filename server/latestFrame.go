package server

import (
	"fmt"
	"sync"
)

type LatestFrame struct {
	frame []byte
	used  bool
	mutex sync.Mutex
}

func (lf *LatestFrame) Get() ([]byte, error) {
	lf.mutex.Lock()
	defer lf.mutex.Unlock()
	if lf.used {
		return nil, fmt.Errorf("frame already sent")
	}

	lf.used = true
	return lf.frame, nil
}
func (lf *LatestFrame) Set(newFrame []byte) {
	lf.mutex.Lock()
	defer lf.mutex.Unlock()
	lf.used = false
	lf.frame = newFrame
}
