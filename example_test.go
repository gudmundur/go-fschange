package fschange_test

import (
	"fmt"
	"time"

	"github.com/gudmundur/go-fschange"
)

func TestExampleNewWatcher() {
	sinceTime := time.Now().Add(-time.Hour * 24)
	w, _ := fschange.NewWatcher(sinceTime)
	defer w.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-w.Events:
				fmt.Println(event)
			case err := <-w.Errors:
				fmt.Println(err)
			}
		}
	}()

	w.Add("../")
	<-done
}
