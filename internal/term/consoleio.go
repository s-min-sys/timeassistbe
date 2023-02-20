package term

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var clear map[string]func()

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func NewConsoleIO() *ConsoleIO {
	return &ConsoleIO{
		reader: bufio.NewReader(os.Stdin),
	}
}

type ConsoleIO struct {
	reader *bufio.Reader
}

func (ci *ConsoleIO) Print(s string) {
	fmt.Print(s)
}

func (ci *ConsoleIO) Println(s string) {
	fmt.Println(s)
}

func (ci *ConsoleIO) Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

func (ci *ConsoleIO) ReadString(tip string) (s string, ok bool, err error) {
	fmt.Print(tip)

	s, err = ci.reader.ReadString('\n')
	if err != nil {
		return
	}

	s = strings.Trim(s, "\r\n\t ")
	if s == "" {
		return
	}

	ok = true

	return
}

func (ci *ConsoleIO) ReadInt(tip string) (n int, ok bool, err error) {
	s, ok, err := ci.ReadString(tip)
	if err != nil || !ok {
		return
	}

	n, err = strconv.Atoi(s)

	return
}
