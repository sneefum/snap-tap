package main

import (
	// "fmt"
	"os"
	"encoding/binary"

	"golang.org/x/sys/unix"
	"github.com/bendahl/uinput"
)

type timeval struct {
	seconds uint64
	microseconds uint64
}

type input_event struct {
	time timeval
	input_type uint16
	keycode uint16
	value uint32
}

var EVIOCGRAB uint = 0x40044590

var keys_pressed []uint16

func main() {
	f, err := os.Open("/dev/input/event12")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	
	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("testkeyboard"))
	if err != nil {
		return
	}
	
	defer keyboard.Close()

	unix.IoctlSetInt(int(f.Fd()), EVIOCGRAB, 1)

	defer unix.IoctlSetInt(int(f.Fd()), EVIOCGRAB, 0)


	buf := make([]byte, 24)

	for {
		f.Read(buf)
		time := timeval{
			seconds: binary.LittleEndian.Uint64(buf[0:8]),
			microseconds: binary.LittleEndian.Uint64(buf[8:16]),
		}
		input := input_event{
			time: time,
			input_type: binary.LittleEndian.Uint16(buf[16:18]),
			keycode: binary.LittleEndian.Uint16(buf[18:20]),
			value: binary.LittleEndian.Uint32(buf[20:24]),
		}

		if input.input_type == 1 {
			switch input.value {
				case 0:
					keyUp(keyboard, input.keycode)
				case 1:
					keyDown(keyboard, input.keycode)
					if input.keycode == 32 { // D
						keyUp(keyboard, 30) // A
					}
					if input.keycode == 30 {
						keyUp(keyboard, 32)
					}
			}
			if input.keycode == 1 { // esc
				return
			}
		}
	}
}

func keyUp(keyboard uinput.Keyboard, keycode uint16) {
	keyboard.KeyUp(int(keycode))
	for i, v := range keys_pressed {
		if v == keycode {
			keys_pressed = append(keys_pressed[:i], keys_pressed[i+1:]...)
			break
		}
	 }
}

func keyDown(keyboard uinput.Keyboard, keycode uint16) {
	keyboard.KeyDown(int(keycode))
	keys_pressed = append(keys_pressed, keycode)
}
