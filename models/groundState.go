package models

import "math"

type GroundState struct {
	State
	Steer     int
	Gas       int
	Brake     int
	Clutch    int
	HandBrake int
	Gear      int
	Aux       [8]bool
}

func (s GroundState) GetType() ControlSchema {
	return s.Schema
}

func (s GroundState) GetBytes() []byte {
	baseMin := -32768
	basMax := 32768
	returnBytes := make([]byte, 7)
	returnBytes[0] = scaleToByte(s.Steer, baseMin, basMax)
	returnBytes[1] = scaleToByte(s.Gas, baseMin, basMax)
	returnBytes[2] = scaleToByte(s.Brake, baseMin, basMax)
	returnBytes[3] = scaleToByte(s.Clutch, baseMin, basMax)
	returnBytes[4] = scaleToByte(s.HandBrake, baseMin, basMax)
	returnBytes[5] = byte(s.Gear)

	var auxMask byte
	for i, buttonOn := range s.Aux {
		if buttonOn {
			auxMask += byte(math.Pow(2, float64(i)))
		}
	}
	returnBytes[6] = auxMask

	return returnBytes
}

func scaleToByte(value int, min int, max int) byte {
	minAllowed := 0
	maxAllowed := 254
	return byte((maxAllowed-minAllowed)*(value-min)/(max-min) + minAllowed)
}
