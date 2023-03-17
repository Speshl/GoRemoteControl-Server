package models

import (
	"math"
)

type GroundState struct {
	State
	Steer     int
	Gas       int
	Brake     int
	Clutch    int
	HandBrake int
	Pan       int
	Tilt      int
	Gear      int
	NumGears  int
	Aux       [8]bool
}

func (s GroundState) GetType() ControlSchema {
	return s.Schema
}

func (s GroundState) GetBytes() []byte {
	baseMin := -32768
	baseMax := 32768
	servoMin := byte(0)
	servoMax := byte(180)
	servoMid := servoMax / 2
	returnBytes := make([]byte, 4)

	returnBytes[0] = mapToRange(s.Steer, baseMin, baseMax, servoMin, servoMax) // steering
	brakeValue := mapToRange(s.Brake*-1, baseMin, baseMax, servoMin, servoMid)
	clutchValue := mapToRange(s.Clutch, baseMin, baseMax, servoMin, servoMax)

	if brakeValue < servoMid {
		returnBytes[1] = brakeValue
	} else if s.Gas > baseMin {
		maxForGear := 90
		switch s.Gear {
		case 1:
			maxForGear = int(servoMid) + 10
		case 2:
			maxForGear = int(servoMid) + 15
		case 3:
			maxForGear = int(servoMid) + 20
		case 4:
			maxForGear = int(servoMid) + 30
		case 5:
			maxForGear = int(servoMid) + 50
		case 6:
			maxForGear = int(servoMax)
		default:
			maxForGear = int(servoMid)
		}

		returnBytes[1] = mapToRange(s.Gas, baseMin, baseMax, servoMid, byte(maxForGear))
	} else {
		returnBytes[1] = servoMid
	}

	if clutchValue > 10 { //clutch overrides
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

func mapToRange(value int, min int, max int, minReturn byte, maxReturn byte) byte {
	return byte(int(maxReturn-minReturn)*(value-min)/(max-min) + int(minReturn))
}
