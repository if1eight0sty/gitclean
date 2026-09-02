package ui

import (
	"fmt"
	"time"
)

func ShowSpinner(message string) func() {
	stop := make(chan bool)
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				fmt.Printf("\r%s Done!          \n", message)
				return
			default:
				fmt.Printf("\r%s %s", spinner[i%len(spinner)], message)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
	return func() {
		stop <- true
	}
}
