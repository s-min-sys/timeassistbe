package main

import (
	"fmt"
	tm "github.com/buger/goterm"
	"time"
)

func NewBox(width, height int) *ConsoleBox {
	return &ConsoleBox{
		box: tm.NewBox(width, height, 0),
	}
}

type ConsoleBox struct {
	box *tm.Box
}

func (cb *ConsoleBox) Print(s string) {
	_, _ = fmt.Fprint(cb.box, s)
}

func (cb *ConsoleBox) Fprintln(s string) {
	_, _ = fmt.Fprintln(cb.box, s)
}

func (cb *ConsoleBox) Printf(format string, a ...any) {
	_, _ = fmt.Fprintf(cb.box, format, a...)
}

func (cb *ConsoleBox) Clear() {
	cb.box.Buf.Reset()
}

func (cb *ConsoleBox) Show(x int, y int) {
	_, _ = tm.Print(tm.MoveTo(cb.box.String(), x, y))
}

func main() {
	tm.Clear()

	box1 := NewBox(30|tm.PCT, 20)

	go func() {
		idx := 0

		for {
			idx++
			box1.Printf("box1:%d\n", idx)
			time.Sleep(time.Second)
		}
	}()

	box2 := NewBox(30|tm.PCT, 20)

	go func() {
		idx := 0

		for {
			idx++
			box2.Printf("box2:%d\n", idx)
			time.Sleep(time.Second * 3)
		}
	}()

	for {
		time.Sleep(time.Millisecond * 100)

		tm.Clear()

		box1.Show(10|tm.PCT, 40|tm.PCT)

		box2.Show(60|tm.PCT, 40|tm.PCT)

		tm.Flush()
	}
}
