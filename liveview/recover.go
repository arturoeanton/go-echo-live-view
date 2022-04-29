package liveview

import "fmt"

func HandleReover() {
	if r := recover(); r != nil {
		fmt.Println("Recovering from panic:", r)
	}
}
