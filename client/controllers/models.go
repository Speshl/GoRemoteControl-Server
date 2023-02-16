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

type AxisEntry struct {
	DeviceID int  `json:"deviceID"`
	AxisID   int  `json:"axisID"`
	Inverted bool `json:"inverted"`
}

type ButtonEntry struct {
	DeviceID int  `json:"deviceID"`
	MaxID    int  `json:"maxID"`
	MinID    *int `json:"minID"`
}

type ConfigEntry struct {
	Axis   *AxisEntry   `json:"axis"`
	Button *ButtonEntry `json:"button"`
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
	Gears     []ButtonEntry `json:"gears"`
	Aux       []ButtonEntry `json:"aux"`
}

type FixedConfig struct {
}

type RotorConfig struct {
}

type QuadConfig struct {
}
