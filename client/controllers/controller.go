package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/0xcafed00d/joystick"
	"github.com/Speshl/GoRemoteControl/models"
	"github.com/tidwall/gjson"
)

var (
	_ ControllerIface = (*Controller)(nil)
)

func CreateController(js []joystick.Joystick, cfgPath string) (*Controller, error) {

	controller := Controller{
		joystick:       js,
		internalStates: make([]joystick.State, len(js)),
	}

	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	schemaValue := gjson.Get(string(byteValue), "schema")
	mappingValue := gjson.Get(string(byteValue), "mapping")

	schema, err := models.ParseControlSchema(schemaValue.String())
	if err != nil {
		return nil, err
	}

	switch schema {
	case models.ControlSchemaGround:
		var cfg GroundConfig
		err = json.Unmarshal([]byte(mappingValue.Raw), &cfg)
		if err != nil {
			return nil, err
		}
		controller.cfg.Config = cfg
	case models.ControlSchemaFixed:
		var cfg FixedConfig
		err = json.Unmarshal([]byte(mappingValue.String()), &cfg)
		if err != nil {
			return nil, err
		}
		controller.cfg.Config = cfg
	case models.ControlSchemaRotor:
		var cfg RotorConfig
		err = json.Unmarshal([]byte(mappingValue.String()), &cfg)
		if err != nil {
			return nil, err
		}
		controller.cfg.Config = cfg
	case models.ControlSchemaQuad:
		var cfg QuadConfig
		err = json.Unmarshal([]byte(mappingValue.String()), &cfg)
		if err != nil {
			return nil, err
		}
		controller.cfg.Config = cfg
	}
	return &controller, nil
}

func (c *Controller) GetUpdatedState() (models.StateIface, error) {
	err := c.UpdateState()
	if err != nil {
		return nil, err
	}

	return c.GetState(c.cfg.Schema)
}

func (c *Controller) UpdateState() error {
	for i, joystick := range c.joystick {
		state, err := joystick.Read()
		if err != nil {
			return err
		}
		c.internalStates[i] = state
	}
	return nil
}

func (c *Controller) GetSchema() models.ControlSchema {
	return c.cfg.Schema
}

func (c *Controller) SetSchema(schema models.ControlSchema) {
	c.cfg.Schema = schema
}

func (c *Controller) GetState(schema models.ControlSchema) (models.StateIface, error) {
	switch schema {
	case models.ControlSchemaGround:
		return c.mapGroundState(), nil
	default:
		return nil, fmt.Errorf("unsupported control schema")
	}
}
