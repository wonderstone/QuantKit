package common

import (
	"testing"
	"time"
)

func TestProcess_GetProgress(t *testing.T) {
	process := NewProcess(1, 50, 100)

	finish := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-finish:
				return
			case <-time.After(100 * time.Millisecond):
				t.Log(process.GetProgress())
			}
		}
	}()

	time.Sleep(10 * time.Second)
	finish <- true
}
