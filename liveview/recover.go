package liveview

import "fmt"

func HandleReover() {
	if r := recover(); r != nil {
		fmt.Println("Recovering from panic:", r)
	}
}

func HandleReoverMsg(msg string) {
	if r := recover(); r != nil {
		fmt.Println(msg, ":", r)
	}
}

func HandleReoverPass() {
	recover()
}
