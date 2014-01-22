package main

import (
	"sync"
	"time"
	"runtime"
	"os"
	"os/signal"
	"fmt"
)

var goroutineStarted = false

func startLoadUpdate() {

	//透過 runtime.NumCPU() 取得 CPU 核心數
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	ch := make(chan int)
	goroutinesRunning := 0

	worker := func() {
		count := 0
		for {
			count += 1
			if count%1000 == 0 {
				time.Sleep(time.Microsecond)
			}
			select {
			case v := <-ch:
				wg.Done()
				fmt.Printf("%d done!\n", v)
				return
			default:
			}
		}
	}
	startGoroutines := func() {
		if goroutineStarted == false {
			goroutineNumToRun := runtime.NumCPU() - goroutinesRunning
			if goroutineNumToRun > 0 {
				for index := 0; index < goroutineNumToRun; index++ {
					wg.Add(1)
					go worker()
					goroutinesRunning += 1
				}
			}
			goroutineStarted = true
		}

	}

	stopGoroutines := func() {
		for index := 0; index < goroutinesRunning; index++ {
			ch<-index
		}
		//close(ch)
		wg.Wait()
	}

	for {
		startGoroutines()
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt, os.Kill)
		select {
		case _ = <-sc:
			stopGoroutines()
			return
		default:
		}
	}
}

func main() {
	startLoadUpdate()
}
