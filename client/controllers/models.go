package controllers

import (
	"github.com/0xcafed00d/joystick"
	"github.com/Speshl/GoRemoteControl/models"
)

type ControllerIface interface {
	UpdateState() error
	GetSchema() models.ControlSchema
	SetSchema(models.ControlSchema)
	GetState(models.ControlSchema) (models.StateIface, error)
	GetUpdatedState() (models.StateIface, error)
}

type Controller struct {
	cfg            Config
	joystick       []joystick.Joystick
	internalStates []joystick.State
	state          models.StateIface
}

type Config struct {
	Schema models.ControlSchema
	Config interface{}
}

type Mapper func([]joystick.State) (models.StateIface, error)

type ConfigEntry struct {
	DeviceID   int              `json:"deviceID"`
	DeviceName string           `json:"deviceName"`
	InputType  models.InputType `json:"inputType"`
	InputID    int              `json:"inputID"`
	Inverted   bool             `json:"inverted"`
}

type GroundConfig struct {
	Steer     ConfigEntry `json:"steer"`
	Gas       ConfigEntry `json:"gas"`
	Brake     ConfigEntry `json:"brake"`
	Clutch    ConfigEntry `json:"clutch"`
	HandBrake ConfigEntry `json:"handbrake"`
	Pan       ConfigEntry `json:"pan"`
	Tilt      ConfigEntry `json:"tilt"`
	NumGears  int
	Gears     []ConfigEntry `json:"gears"`
	Aux       []ConfigEntry `json:"aux"`
}

type FixedConfig struct {
}

type RotorConfig struct {
}

type QuadConfig struct {
}
