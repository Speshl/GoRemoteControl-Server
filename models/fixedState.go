package models

type FixedState struct {
	State
}

func (s FixedState) GetType() ControlSchema {
	return s.Schema
}

func (s FixedState) GetBytes() []byte {
	return nil
}
