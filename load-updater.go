package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

var sc_clk_tck C.long

var goroutineStarted = false

type CPUInfo map[string]float64

func calculate(t1, t2 CPUInfo) float64 {
	sum := func(data CPUInfo) float64 {
		all := 0.0
		for _, value := range data {
			all += value
		}
		return all
	}
	t1All := sum(t1)
	t1Busy := t1All - t1["idle"]

	t2All := sum(t2)
	t2Busy := t2All - t2["idle"]

	if t2Busy <= t1Busy {
		return 0.0
	}
	busyDelta := t2Busy - t1Busy
	allDelta := t2All - t1All
	busyPerc := (busyDelta / allDelta) * 100
	return busyPerc
}

func getSystemPerCPUTimes() []CPUInfo {
	statInfo, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		fmt.Println("Can not fetch /proc/stat")
		return nil
	}
	lines := strings.Split(string(statInfo), "\n")
	results := make([]CPUInfo, 16)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.HasPrefix(fields[0], "cpu") {
			var oneCPU = make(CPUInfo)
			userPart, _ := strconv.ParseFloat(fields[1], 64)
			oneCPU["user"] = userPart / float64(sc_clk_tck)
			nicePart, _ := strconv.ParseFloat(fields[2], 64)
			oneCPU["nice"] = nicePart / float64(sc_clk_tck)
			systemPart, _ := strconv.ParseFloat(fields[3], 64)
			oneCPU["system"] = systemPart / float64(sc_clk_tck)
			idlePart, _ := strconv.ParseFloat(fields[4], 64)
			oneCPU["idle"] = idlePart / float64(sc_clk_tck)
			iowaitPart, _ := strconv.ParseFloat(fields[5], 64)
			oneCPU["iowait"] = iowaitPart / float64(sc_clk_tck)
			irqPart, _ := strconv.ParseFloat(fields[6], 64)
			oneCPU["irq"] = irqPart / float64(sc_clk_tck)
			softirqPart, _ := strconv.ParseFloat(fields[7], 64)
			oneCPU["softirq"] = softirqPart / float64(sc_clk_tck)
			results = append(results, oneCPU)
		}
	}
	return results
}

func cpuPercent(interval int) []float64 {
	blocking := true
	if interval <= 0 {
		blocking = false
	}
	var t1a []CPUInfo
	if blocking {
		t1a = getSystemPerCPUTimes()
		time.Sleep(time.Second * time.Duration(interval))
	}
	t2a := getSystemPerCPUTimes()
	infoPartNum := len(t2a)
	ret := make([]float64, 16)
	for index := 0; index < infoPartNum; index++ {
		ret = append(ret, calculate(t1a[index], t2a[index]))
	}
	return ret
}

func sumFloat64(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum
}

func genRandInt() int {
	seed := time.Now().Unix()
	r := rand.New(rand.NewSource(seed))
	return r.Intn(50)
}

func startLoadUpdate(pauseInterval, runDuration, goroutinesPerCPU, loadLimit int, forever bool) {

	//透過 runtime.NumCPU() 取得 CPU 核心數
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	ch := make(chan int)
	exitNotify := make(chan bool)
	goroutinesRunning := 0

	worker := func(goodLuck bool) {
		count := 0
		for {
			count += 1
			if count%pauseInterval == 0 {
				time.Sleep(time.Microsecond)
			}
			select {
			case <-ch:
				exitNotify <- true
				runtime.Goexit()
			default:
				_ = float64(count+500) / 10.22
				if goodLuck {
					_ = float64(count+1000) / 10.22
					_ = float64(count+5000) / 10.22
				}
			}
		}
	}

	stopGoroutines := func() {
		// fmt.Printf("before, goroutinesRunning: %d\n", goroutinesRunning)
		for index := 0; index < goroutinesRunning; index++ {
			ch <- index
		}
		goroutinesRunningBackUp := goroutinesRunning
		for index := 0; index < goroutinesRunningBackUp; index++ {
			// fmt.Printf("exitNotify: %t\n", <-exitNotify)
			<-exitNotify
			goroutinesRunning -= 1
		}
		goroutineStarted = false
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	forTimeOut := time.After(time.Duration(runDuration) * time.Minute)
	for {
		select {
		case <-sc:
			stopGoroutines()
			return
		case now := <-forTimeOut:
			stopGoroutines()
			if forever {
				time.Sleep(time.Duration(genRandInt()*30) * time.Second)
			} else {
				fmt.Printf("%s, 运行时间到！\n", now.Format("2006-01-02 15:04:05"))
				return
			}
		default:
			// 为了不阻塞，得加default分支
		}

		if goroutineStarted == false {
			goroutineNumToRun := runtime.NumCPU()*goroutinesPerCPU - goroutinesRunning
			goodLuck := false
			if genRandInt() >= 25 {
				goodLuck = true
			}
			if goroutineNumToRun > 0 {
				for index := 0; index < goroutineNumToRun; index++ {
					go worker(goodLuck)
					goroutinesRunning += 1
					// fmt.Printf("goroutinesRunning: %d\n", goroutinesRunning)
				}
			}
			goroutineStarted = true
		}
		loads := cpuPercent(5)

		if sumFloat64(loads) >= float64((runtime.NumCPU()+1)*loadLimit) {
			// fmt.Println(sumFloat64(loads))
			stopGoroutines()
		}
		time.Sleep(time.Duration(5) * time.Second)
	}

}

var (
	pauseInterval    = flag.Int("i", 500000, "can cause the server to have different CPU load")
	runDuration      = flag.Int("d", 10, "time duration this program to run, whose unit is minute")
	forever          = flag.Bool("e", true, "run it forever")
	goroutinesPerCPU = flag.Int("n", 1, "number of goroutine to run on one CPU")
	loadLimit        = flag.Int("l", 50, "cpu usage")
)

func main() {
	flag.Parse()
	sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
	startLoadUpdate(*pauseInterval, *runDuration, *goroutinesPerCPU, *loadLimit, *forever)
}
