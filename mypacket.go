package main

import (
	"fmt"
	"time"
)

func pinger(key string, c chan string) {
	for i := 0; ; i++ {
		c <- key + "ping"
	}
}
func printer(c chan string) {
	for {
		msg := <-c
		fmt.Println(msg)
		time.Sleep(time.Second * 1)
	}
}
func ponger(key string, c chan string) {
	for i := 0; ; i++ {
		c <- key + "pong"
	}
}
func main() {
	var c chan string = make(chan string)

	go pinger("Kate ", c)
	go pinger("Marita ", c)
	go ponger("Kate ", c)
	go ponger("Marita ", c)
	go printer(c)

	var input string
	fmt.Scanln(&input)
}
