package main

import (
	"fmt"
	"os"
	"encoding/binary"

	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"
	"github.com/bendahl/uinput"
)

type Config struct {
	Combinations [][2]int `yaml:"combinations"`
}

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

func findPair(c *Config, x int) *int {
	for _, row := range c.Combinations {
		if row[0] == x {
			return &row[1]
		}
		if row[1] == x {
			return &row[0]
		}
	}
	return nil
}

func main() {
	config_data, err := os.ReadFile("./config.yml")
	if err != nil {
		panic(err)
	}

	config := Config{}
	err = yaml.Unmarshal([]byte(config_data), &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded config.yml!\n")

	f, err := os.Open("/dev/input/by-path/pci-0000:01:00.0-usb-0:9:1.0-event-kbd")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Printf("Opened keyboard device!\n")
	
	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("testkeyboard"))
	if err != nil {
		return
	}

	fmt.Printf("Created uinput device!\n")
	
	defer keyboard.Close()

	unix.IoctlSetInt(int(f.Fd()), EVIOCGRAB, 1)
	
	defer unix.IoctlSetInt(int(f.Fd()), EVIOCGRAB, 0)

	buf := make([]byte, 24)

	fmt.Printf("Made input buffer!\n")

	for {
		n, err := f.Read(buf)

		if err != nil {
        	panic(err)
    	}

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

		if n != 24 {
        	fmt.Println("Partial read:", n)
        	continue
   		}

    	fmt.Println("Event bytes:", buf)

		if input.input_type == 1 {
			switch input.value {
				case 0:
					keyUp(keyboard, input.keycode)
				case 1:
					keyDown(keyboard, input.keycode)
					keyUp(keyboard, uint16(*findPair(&config, int(input.keycode))))
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
