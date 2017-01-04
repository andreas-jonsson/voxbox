// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package platform

const (
	KeyUnknown = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEsc
	KeyReturn
)

const (
	MouseButtonDown = iota
	MouseButtonUp
	MouseWheel
)

type MouseState struct {
	X, Y    int
	Buttons [3]bool
}

type (
	Event     interface{}
	QuitEvent struct{}

	KeyUpEvent struct {
		Rune rune
		Key  int
	}

	KeyDownEvent KeyUpEvent

	MouseWheelEvent struct {
		X, Y int
	}

	MouseMotionEvent struct {
		X, Y, XRel, YRel int
	}

	MouseButtonEvent struct {
		X, Y, Button, Type int
	}
)
