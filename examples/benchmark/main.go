package main

import (
	"fmt"
	"time"

	"github.com/headblockhead/pio"
	"github.com/headblockhead/pio/simulation"
)

func main() {
	s := simulation.New()

	for i := range 10 {
		r := pio.NewRP2040(fmt.Sprintf("RP2040 %d", i))
		_ = s.AddTicker(r)
	}

	var end bool = false
	var timer int64 = 5
	go func() {
		time.Sleep(time.Duration(timer) * time.Second)
		end = true
	}()
	var i int64
	for {
		_ = s.Tick()
		i++
		if end {
			break
		}
	}

	fmt.Printf("%d ticks over %d seconds: %d ticks/second", i, timer, i/timer)
}
