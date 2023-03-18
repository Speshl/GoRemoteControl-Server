package models

import (
	"math"
)

type GroundState struct {
	State
	Steer          int
	Gas            int
	Brake          int
	Clutch         int
	HandBrake      int
	Pan            int
	Tilt           int
	Gear           int
	NumGears       int
	Aux            [8]bool
	InvertSteering bool
	InvertEsc      bool
}

func (s GroundState) GetType() ControlSchema {
	return s.Schema
}

func (s GroundState) GetBytes() []byte {
	baseMin := -32768
	baseMax := 32768
	servoMin := 0
	servoMax := 180
	servoMid := servoMax / 2
	returnBytes := make([]byte, 4)

	returnBytes[0] = mapToRange(s.Steer, baseMin, baseMax, servoMin, servoMax) // steering
	if s.InvertSteering {
		returnBytes[0] = mapToRange(s.Steer*-1, baseMin, baseMax, servoMin, servoMax) // steering
	}

	offsetPerGear := 90 / s.NumGears
	gearOffset := offsetPerGear * s.Gear
	if s.Gear == s.NumGears {
		gearOffset = 90
	}

	gasValue := mapToRange(s.Gas, baseMin, baseMax, servoMid, servoMid+gearOffset)
	brakeValue := mapToRange(s.Brake*-1, baseMin, baseMax, servoMin, servoMid)
	clutchValue := mapToRange(s.Clutch, baseMin, baseMax, servoMin, servoMax)
	if s.InvertEsc {
		gasValue = mapToRange(s.Gas*-1, baseMin, baseMax, servoMid-gearOffset, servoMid)
		brakeValue = mapToRange(s.Brake, baseMin, baseMax, servoMid, servoMax)
	}

	if brakeValue != byte(servoMid) {
		returnBytes[1] = brakeValue
	} else if clutchValue > byte(servoMid) {
		returnBytes[1] = byte(servoMid)
	} else if gasValue != byte(servoMid) {
		returnBytes[1] = gasValue
	} else {
		returnBytes[1] = byte(servoMid)
	}

	panValue := mapToRange(s.Pan, baseMin, baseMax, 0, 15)
	tiltValue := mapToRange(s.Tilt, baseMin, baseMax, 0, 15)
	panAndTilt := (panValue << 4) | tiltValue
	if panAndTilt > 255 {
		returnBytes[2] = 255
	} else {
		returnBytes[2] = panAndTilt
	}

	var auxMask byte
	for i, buttonOn := range s.Aux {
		if buttonOn {
			auxMask += byte(math.Pow(2, float64(i)))
		}
	}
	returnBytes[3] = auxMask
	//log.Printf("State: %+v\n", s)
	//log.Printf("StateBytes: %+v\n", returnBytes)
	return returnBytes
}

func mapToRange(value int, min int, max int, minReturn int, maxReturn int) byte {
	return byte(int(maxReturn-minReturn)*(value-min)/(max-min) + int(minReturn))
}
