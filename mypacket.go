package main

import (
	"fmt"
	"time"
)

type Girl struct {
	name   string
	number int
}

func pinger(c chan Girl) {
	for i := 0; ; i++ {
		c <- "ping"
	}
}
func printer(c chan string) {
	for {
		msg := <-c
		fmt.Println(msg)
		time.Sleep(time.Second * 1)
	}
}
func ponger(c chan string) {
	for i := 0; ; i++ {
		c <- "pong"
	}
}
func main() {
	g := Girl{"Kate", 0}
	var c chan Girl = make(chan Girl)

	go pinger(c)
	go ponger(c)
	go printer(c)

	var input string
	fmt.Scanln(&input)
}
