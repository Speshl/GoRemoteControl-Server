package controllers

import (
	"fmt"
	"math"

	"github.com/0xcafed00d/joystick"
	"github.com/Speshl/GoRemoteControl/models"
)

/*
	func getTestG27Config() GroundConfig {
		return GroundConfig{
			Steer:     ConfigEntry{DeviceID: 0, InputType: models.InputTypeAxis, InputID: 0},
			Gas:       ConfigEntry{DeviceID: 0, InputType: models.InputTypeAxis, InputID: 2, Inverted: true},
			Brake:     ConfigEntry{DeviceID: 0, InputType: models.InputTypeAxis, InputID: 3, Inverted: true},
			Clutch:    ConfigEntry{DeviceID: 0, InputType: models.InputTypeAxis, InputID: 4, Inverted: true},
			HandBrake: ConfigEntry{DeviceID: 0, InputType: models.InputTypeAxis, InputID: 5},
			Gears: []ConfigEntry{
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 8},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 9},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 10},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 11},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 12},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 13},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 14},
			},
			Aux: []ConfigEntry{
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 1},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 2},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 3},
				{DeviceID: 0, InputType: models.InputTypeButton, InputID: 4},
			},
		}
	}
*/
func (c *Controller) mapGroundState() (models.StateIface, error) {
	//cfg := getTestG27Config()
	states := c.internalStates
	cfg := c.cfg.Config.(GroundConfig)
	err := c.validateGroundConfig(cfg, states)
	if err != nil {
		return nil, err
	}
	returnState := models.GroundState{}

	returnState.Steer = states[cfg.Steer.DeviceID].AxisData[cfg.Steer.InputID]
	if cfg.Steer.Inverted {
		returnState.Steer = returnState.Steer * -1
	}

	returnState.Gas = states[cfg.Gas.DeviceID].AxisData[cfg.Gas.InputID]
	if cfg.Gas.Inverted {
		returnState.Gas = returnState.Gas * -1
	}

	returnState.Brake = states[cfg.Brake.DeviceID].AxisData[cfg.Brake.InputID]
	if cfg.Brake.Inverted {
		returnState.Brake = returnState.Brake * -1
	}

	returnState.Clutch = states[cfg.Clutch.DeviceID].AxisData[cfg.Clutch.InputID]
	if cfg.Clutch.Inverted {
		returnState.Clutch = returnState.Clutch * -1
	}

	returnState.HandBrake = states[cfg.HandBrake.DeviceID].AxisData[cfg.HandBrake.InputID]
	if cfg.HandBrake.Inverted {
		returnState.HandBrake = returnState.HandBrake * -1
	}

	returnState.Pan = states[cfg.Pan.DeviceID].AxisData[cfg.Pan.InputID]
	if cfg.Pan.Inverted {
		returnState.Pan = returnState.Pan * -1
	}

	returnState.Tilt = states[cfg.Tilt.DeviceID].AxisData[cfg.Tilt.InputID]
	if cfg.Tilt.Inverted {
		returnState.Tilt = returnState.Tilt * -1
	}

	returnState.NumGears = len(cfg.Gears) - 1 //Don't count reverse

	for gear, gearButton := range cfg.Gears {
		bitValue := uint32(math.Pow(2, float64(gearButton.InputID)))
		//fmt.Printf("BitValue: %d\n", bitValue)
		//fmt.Printf(" %d & %d = %d\n", state.Buttons, bitValue, state.Buttons&bitValue)
		if gearButton.Inverted {
			if states[cfg.Gas.DeviceID].Buttons&bitValue <= 0 {
				if gear == len(cfg.Gears)-1 {
					returnState.Gear = -1
				} else {
					returnState.Gear = gear + 1
				}
			}
		} else {
			if states[cfg.Gas.DeviceID].Buttons&bitValue > 0 {
				if gear == len(cfg.Gears)-1 {
					returnState.Gear = -1
				} else {
					returnState.Gear = gear + 1
				}
			}
		}
	}

	for pos, auxButton := range cfg.Aux {
		bitValue := uint32(math.Pow(2, float64(auxButton.InputID)))
		if auxButton.Inverted {
			if states[auxButton.DeviceID].Buttons&bitValue <= 0 {
				returnState.Aux[pos] = true
			}
		} else {
			if states[auxButton.DeviceID].Buttons&bitValue > 0 {
				returnState.Aux[pos] = true
			}
		}
	}

	return returnState, nil
}

func (c *Controller) validateGroundConfig(cfg GroundConfig, states []joystick.State) error {
	numJoysticks := len(states)

	if cfg.Steer.DeviceID >= numJoysticks {
		return fmt.Errorf("not enough joysticks connected - %+v", cfg.Steer)
	}
	if cfg.Steer.InputType != models.InputTypeAxis {
		return fmt.Errorf("steering input type must an axis - %+v", cfg.Steer)
	}
	if cfg.Steer.InputID >= len(states[cfg.Steer.DeviceID].AxisData) {
		return fmt.Errorf("device does not have enough axis - %+v", cfg.Steer)
	}

	if cfg.Gas.DeviceID >= numJoysticks {
		return fmt.Errorf("not enough joysticks connected - %+v", cfg.Gas)
	}
	if cfg.Gas.InputType != models.InputTypeAxis {
		return fmt.Errorf("gas input type must an axis - %+v", cfg.Gas)
	}
	if cfg.Gas.InputID >= len(states[cfg.Gas.DeviceID].AxisData) {
		return fmt.Errorf("device does not have enough axis - %+v", cfg.Gas)
	}

	if cfg.Brake.DeviceID >= numJoysticks {
		return fmt.Errorf("not enough joysticks connected - %+v", cfg.Brake)
	}
	if cfg.Brake.InputType != models.InputTypeAxis {
		return fmt.Errorf("brake input type must an axis - %+v", cfg.Brake)
	}
	if cfg.Brake.InputID >= len(states[cfg.Brake.DeviceID].AxisData) {
		return fmt.Errorf("device does not have enough axis - %+v", cfg.Brake)
	}

	if cfg.Clutch.DeviceID >= numJoysticks {
		return fmt.Errorf("not enough joysticks connected - %+v", cfg.Clutch)
	}
	if cfg.Clutch.InputType != models.InputTypeAxis {
		return fmt.Errorf("clutch input type must an axis - %+v", cfg.Clutch)
	}
	if cfg.Clutch.InputID >= len(states[cfg.Clutch.DeviceID].AxisData) {
		return fmt.Errorf("device does not have enough axis - %+v", cfg.Clutch)
	}

	if cfg.HandBrake.DeviceID >= numJoysticks {
		return fmt.Errorf("not enough joysticks connected - %+v", cfg.HandBrake)
	}
	if cfg.HandBrake.InputType != models.InputTypeAxis {
		return fmt.Errorf("handbrake input type must an axis - %+v", cfg.HandBrake)
	}
	if cfg.HandBrake.InputID >= len(states[cfg.HandBrake.DeviceID].AxisData) {
		return fmt.Errorf("device does not have enough axis - %+v", cfg.HandBrake)
	}

	return nil
}
