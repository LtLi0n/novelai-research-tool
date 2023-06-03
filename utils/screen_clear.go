package utils

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
)

var clear map[string]func() //map for screen clear funcs

func init_clear() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear") // Linux
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") // Windows
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func TryScreenClear() error {
	if len(clear) == 0 {
		init_clear()
	}
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
		return nil
	}
	return errors.New("Couldn't clear screen")
}
