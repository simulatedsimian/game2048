// +build windows

package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

const (
	_MAXPNAMELEN            = 32
	_MAX_JOYSTICKOEMVXDNAME = 260

	JOY_RETURNX        = 1
	JOY_RETURNY        = 2
	JOY_RETURNZ        = 4
	JOY_RETURNR        = 8
	JOY_RETURNU        = 16
	JOY_RETURNV        = 32
	JOY_RETURNPOV      = 64
	JOY_RETURNBUTTONS  = 128
	JOY_RETURNRAWDATA  = 256
	JOY_RETURNPOVCTS   = 512
	JOY_RETURNCENTERED = 1024
	JOY_USEDEADZONE    = 2048
	JOY_RETURNALL      = (JOY_RETURNX | JOY_RETURNY | JOY_RETURNZ | JOY_RETURNR | JOY_RETURNU | JOY_RETURNV | JOY_RETURNPOV | JOY_RETURNBUTTONS)
)

type JOYCAPS struct {
	wMid        uint16
	wPid        uint16
	szPname     [_MAXPNAMELEN]uint16
	wXmin       uint32
	wXmax       uint32
	wYmin       uint32
	wYmax       uint32
	wZmin       uint32
	wZmax       uint32
	wNumButtons uint32
	wPeriodMin  uint32
	wPeriodMax  uint32
	wRmin       uint32
	wRmax       uint32
	wUmin       uint32
	wUmax       uint32
	wVmin       uint32
	wVmax       uint32
	wCaps       uint32
	wMaxAxes    uint32
	wNumAxes    uint32
	wMaxButtons uint32
	szRegKey    [_MAXPNAMELEN]uint16
	szOEMVxD    [_MAX_JOYSTICKOEMVXDNAME]uint16
}

type JOYINFOEX struct {
	dwSize         uint32
	dwFlags        uint32
	dwAxis         [6]uint32
	dwButtons      uint32
	dwButtonNumber uint32
	dwPOV          uint32
	dwReserved1    uint32
	dwReserved2    uint32
}

var (
	winmmdll      = windows.MustLoadDLL("Winmm.dll")
	joyGetPosEx   = winmmdll.MustFindProc("joyGetPosEx")
	joyGetDevCaps = winmmdll.MustFindProc("joyGetDevCapsW")
)

type JoystickImpl struct {
	id          int
	axisCount   int
	buttonCount int
	name        string
	state       JoystickInfo
}

func OpenJoystick(id int) (Joystick, error) {

	js := &JoystickImpl{}
	js.id = id

	err := js.getJoyCaps()
	if err == nil {
		return js, nil
	}
	return nil, err
}

func (js *JoystickImpl) getJoyCaps() error {
	var caps JOYCAPS
	ret, _, _ := joyGetDevCaps.Call(uintptr(js.id), uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps))

	if ret != 0 {
		return fmt.Errorf("Failed to read Joystick %d", js.id)
	} else {
		js.axisCount = int(caps.wNumAxes)
		js.buttonCount = int(caps.wNumButtons)
		js.name = windows.UTF16ToString(caps.szPname[:])
		return nil
	}
}

func (js *JoystickImpl) getJoyPosEx() error {
	var info JOYINFOEX
	info.dwSize = uint32(unsafe.Sizeof(info))
	info.dwFlags = JOY_RETURNALL
	ret, _, _ := joyGetPosEx.Call(uintptr(js.id), uintptr(unsafe.Pointer(&info)))

	if ret != 0 {
		return fmt.Errorf("Failed to read Joystick %d", js.id)
	} else {
		js.state.Buttons = info.dwButtons

		for i, v := range info.dwAxis {
			js.state.AxisData[i] = int(v)
		}
		return nil
	}
}

func (js *JoystickImpl) AxisCount() int {
	return js.axisCount
}

func (js *JoystickImpl) ButtonCount() int {
	return js.buttonCount
}

func (js *JoystickImpl) Name() string {
	return js.name
}

func (js *JoystickImpl) Read() JoystickInfo {
	js.getJoyPosEx()
	return js.state
}

func (js *JoystickImpl) Close() {
	// no impl under windows
}