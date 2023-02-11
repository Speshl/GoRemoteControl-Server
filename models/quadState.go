package models

type QuadState struct {
	State
}

func (s QuadState) GetType() ControlSchema {
	return s.Schema
}

func (s QuadState) GetBytes() []byte {
	return nil
}
