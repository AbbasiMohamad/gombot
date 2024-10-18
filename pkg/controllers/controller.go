package controllers

import "time"

const (
	SleepTime = 2 * time.Second
)

func sleep() {
	time.Sleep(SleepTime)
}
