package models

import "time"

//go:generate go-enum --marshal

// ENUM(ground, fixed, rotor, quad)
type ControlSchema int

type StateIface interface {
	GetType() ControlSchema
	GetBytes() []byte
}

type State struct {
	Schema ControlSchema
}

type Packet struct {
	StateType ControlSchema
	State     StateIface
	SentAt    time.Time
}
