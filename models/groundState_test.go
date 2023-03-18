package models_test

import (
	"testing"

	"github.com/Speshl/GoRemoteControl_Server/models"

	. "github.com/smartystreets/goconvey/convey"
)

type MapTestCase struct {
	State            models.GroundState
	ExpectedESc      int
	ExpectedSteering int
}

func TestMap(t *testing.T) {
	testCases := []MapTestCase{
		{
			State: GroundState{ //All pedals full, steering full right
				Gas:            32768,
				Steer:          32768,
				Brake:          32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      false,
			},
			ExpectedESc:      0,
			ExpectedSteering: 180,
		},
		{
			State: GroundState{ //No Input
				Gas:            -32768,
				Steer:          0,
				Brake:          -32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      false,
			},
			ExpectedESc:      90,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //half throttle in 6th gear
				Gas:            0,
				Steer:          0,
				Brake:          -32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      false,
			},
			ExpectedESc:      135,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //half throttle in 1st
				Gas:            0,
				Steer:          0,
				Brake:          -32768,
				Gear:           1,
				InvertSteering: false,
				InvertEsc:      false,
			},
			ExpectedESc:      97,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //full throttle in 1st
				Gas:            32768,
				Steer:          0,
				Brake:          -32768,
				Gear:           1,
				InvertSteering: false,
				InvertEsc:      false,
			},
			ExpectedESc:      105,
			ExpectedSteering: 90,
		},

		//inverted esc
		{
			State: GroundState{ //All pedals full, steering full right
				Gas:            32768,
				Steer:          32768,
				Brake:          32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      true,
			},
			ExpectedESc:      180,
			ExpectedSteering: 180,
		},
		{
			State: GroundState{ //No Input
				Gas:            -32768,
				Steer:          0,
				Brake:          -32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      true,
			},
			ExpectedESc:      90,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //half throttle in 6th gear
				Gas:            0,
				Steer:          0,
				Brake:          -32768,
				Gear:           6,
				InvertSteering: false,
				InvertEsc:      true,
			},
			ExpectedESc:      45,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //half throttle in 1st
				Gas:            0,
				Steer:          0,
				Brake:          -32768,
				Gear:           1,
				InvertSteering: false,
				InvertEsc:      true,
			},
			ExpectedESc:      82,
			ExpectedSteering: 90,
		},
		{
			State: GroundState{ //full throttle in 1st
				Gas:            32768,
				Steer:          0,
				Brake:          -32768,
				Gear:           1,
				InvertSteering: false,
				InvertEsc:      true,
			},
			ExpectedESc:      75,
			ExpectedSteering: 90,
		},
	}

	for _, testCase := range testCases {
		Convey("Given a test state", t, func(c C) {
			bytes := testCase.State.GetBytes()
			So(bytes[0], ShouldEqual, testCase.ExpectedSteering)
			So(bytes[1], ShouldEqual, testCase.ExpectedESc)
		})
	}
}
