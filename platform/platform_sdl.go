// +build !mobile

// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package platform

import (
	"os"
	"os/user"
	"path"
	"runtime"

	"github.com/veandco/go-sdl2/sdl"
)

var keyMapping = map[sdl.Keycode]int{
	sdl.K_UP:     KeyUp,
	sdl.K_DOWN:   KeyDown,
	sdl.K_LEFT:   KeyLeft,
	sdl.K_RIGHT:  KeyRight,
	sdl.K_ESCAPE: KeyEsc,
	sdl.K_RETURN: KeyReturn,
}

var mouseMapping = map[int]int{
	sdl.MOUSEBUTTONDOWN: MouseButtonDown,
	sdl.MOUSEBUTTONUP:   MouseButtonUp,
	sdl.MOUSEWHEEL:      MouseWheel,
}

func init() {
	runtime.LockOSThread()

	if runtime.GOOS == "windows" {
		ConfigPath = path.Join(os.Getenv("LOCALAPPDATA"), "Voxbox")
	} else {
		if usr, err := user.Current(); err == nil {
			ConfigPath = path.Join(usr.HomeDir, ".config", "voxbox")
		}
	}

	ConfigPath = path.Clean(ConfigPath)
	os.MkdirAll(ConfigPath, 0755)
}

func Init() error {
	idCounter = 0
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_GAMECONTROLLER); err != nil {
		return err
	}
	return nil
}

func Shutdown() {
	sdl.Quit()
}

func Mouse() MouseState {
	x, y, buttons := sdl.GetMouseState()

	window := sdl.GetMouseFocus()
	if window == nil {
		return MouseState{}
	}

	left := (buttons & sdl.ButtonLMask()) != 0
	middle := (buttons & sdl.ButtonMMask()) != 0
	right := (buttons & sdl.ButtonRMask()) != 0

	return MouseState{X: x, Y: y, Buttons: [3]bool{left, middle, right}}
}

func PollEvent() Event {
	event := sdl.PollEvent()
	if event == nil {
		return nil
	}

	switch t := event.(type) {
	case *sdl.QuitEvent:
		return &QuitEvent{}
	case *sdl.KeyUpEvent:
		ev := &KeyUpEvent{}
		if key, ok := keyMapping[t.Keysym.Sym]; ok {
			ev.Key = key
			ev.Rune = rune(t.Keysym.Unicode)
		} else {
			ev.Key = KeyUnknown
		}
		return ev
	case *sdl.KeyDownEvent:
		ev := &KeyDownEvent{}
		if key, ok := keyMapping[t.Keysym.Sym]; ok {
			ev.Key = key
			ev.Rune = rune(t.Keysym.Unicode)
		} else {
			ev.Key = KeyUnknown
		}
		return ev
	case *sdl.MouseButtonEvent:
		ev := &MouseButtonEvent{
			Button: int(t.Button),
			X:      int(t.X),
			Y:      int(t.Y),
		}

		switch t.Type {
		case sdl.MOUSEBUTTONDOWN:
			ev.Type = MouseButtonDown
		case sdl.MOUSEBUTTONUP:
			ev.Type = MouseButtonUp
		case sdl.MOUSEWHEEL:
			ev.Type = MouseWheel
		}
		return ev
	case *sdl.MouseMotionEvent:
		return &MouseMotionEvent{
			X:    int(t.X),
			Y:    int(t.Y),
			XRel: int(t.XRel),
			YRel: int(t.YRel),
		}
	case *sdl.MouseWheelEvent:
		return &MouseWheelEvent{
			X: int(t.X),
			Y: int(t.Y),
		}
	}

	return nil
}
