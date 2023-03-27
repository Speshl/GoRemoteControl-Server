package server

import (
	"fmt"
	"sync"

	"github.com/Speshl/GoRemoteControl_Server/models"
)

type LatestState struct {
	state models.StateIface
	used  bool
	mutex sync.RWMutex
}

func (ls *LatestState) Get() (models.StateIface, error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()
	if ls.used {
		return nil, fmt.Errorf("state already sent")
	}

	ls.used = true
	return ls.state, nil
}
func (ls *LatestState) Set(newState models.StateIface) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.used = false
	ls.state = newState
}
