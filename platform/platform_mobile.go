// +build mobile

// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package platform

import (
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
)

const maxEvents = 128

var (
	InputEventChan = make(chan interface{}, maxEvents)
	sizeEvent      size.Event
)

func Init() error {
	idCounter = 0
	return nil
}

func Shutdown() {
}

func Mouse() MouseState {
	return MouseState{}
}

func PollEvent() Event {
	select {
	case ev, ok := <-InputEventChan:
		if ok {
			switch e := ev.(type) {
			case size.Event:
				sizeEvent = e
			case touch.Event:
				ws := 320 / float32(sizeEvent.WidthPx)
				hs := 200 / float32(sizeEvent.HeightPx)

				if e.Type == touch.TypeBegin {
					return &MouseButtonEvent{X: int(e.X * ws), Y: int(e.Y * hs), Button: 0, Type: MouseButtonDown}
				} else if e.Type == touch.TypeEnd {
					return &MouseButtonEvent{X: int(e.X * ws), Y: int(e.Y * hs), Button: 0, Type: MouseButtonUp}
				} else {
					return &MouseMotionEvent{X: int(e.X * ws), Y: int(e.Y * hs)}
				}
			}
		} else {
			return QuitEvent{}
		}
	default:
	}
	return nil
}
