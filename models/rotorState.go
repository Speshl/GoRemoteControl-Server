package models

type RotorState struct {
	State
}

func (s RotorState) GetType() ControlSchema {
	return s.Schema
}

func (s RotorState) GetBytes() []byte {
	return nil
}
