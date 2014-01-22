package main

import (
	"sync"
	"time"
	"runtime"
	"os"
	"os/signal"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
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

/*
def get_system_per_cpu_times():
    """Return a list of namedtuple representing the CPU times
    for every CPU available on the system.
    """
    cpus = []
    f = open('/proc/stat', 'r')
    # get rid of the first line who refers to system wide CPU stats
    try:
        f.readline()
        for line in f.readlines():
            if line.startswith('cpu'):
                values = line.split()[1:8]
                values = tuple([float(x) / _CLOCK_TICKS for x in values])
                entry = nt_sys_cputimes(*values[:7])
                cpus.append(entry)
        return cpus
    finally:
        f.close()
*/

type CPUInfo map[string]float64

func calculate(t1, t2 CPUInfo) float64 {
	sum := func(data CPUInfo) float64 {
		all := 0.0
		for _, value := range (data) {
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
	for _, line := range (lines) {
		fields := strings.Fields(line)
		fmt.Printf("fields length: %d\n", len(fields))
		if len(fields) > 0 && strings.HasPrefix(fields[0], "cpu") {
			var oneCPU = make(CPUInfo)
			userPart, _ := strconv.ParseFloat(fields[1], 64)
			oneCPU["user"] = userPart/float64(sc_clk_tck)
			nicePart, _ := strconv.ParseFloat(fields[2], 64)
			oneCPU["nice"] = nicePart/float64(sc_clk_tck)
			systemPart, _ := strconv.ParseFloat(fields[3], 64)
			oneCPU["system"] = systemPart/float64(sc_clk_tck)
			idlePart, _ := strconv.ParseFloat(fields[4], 64)
			oneCPU["idle"] = idlePart/float64(sc_clk_tck)
			iowaitPart, _ := strconv.ParseFloat(fields[5], 64)
			oneCPU["iowait"] = iowaitPart/float64(sc_clk_tck)
			irqPart, _ := strconv.ParseFloat(fields[6], 64)
			oneCPU["irq"] = irqPart/float64(sc_clk_tck)
			softirqPart, _ := strconv.ParseFloat(fields[7], 64)
			oneCPU["softirq"] = softirqPart/float64(sc_clk_tck)
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
	for _, value := range (values) {
		sum += value
	}
	return sum
}

func startLoadUpdate() {

	//透過 runtime.NumCPU() 取得 CPU 核心數
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	ch := make(chan int)
	worker := func() {
		count := 0
		for {
			count += 1
			if count%24000 == 0 {
				time.Sleep(time.Microsecond)
			}
			select {
			case _ = <-ch:
				wg.Done()
				// 退出当前执行的goroutine，但是defer函数还会继续调用
				runtime.Goexit()
			default:
				_ = float64(count)/10.22
			}
		}
	}
	startGoroutines := func() {
		if goroutineStarted == false {
			// 返回正在执行和排队的任务总数
			goroutinesRunning := runtime.NumGoroutine()
			goroutineNumToRun := runtime.NumCPU() - goroutinesRunning
			if goroutineNumToRun > 0 {
				for index := 0; index < goroutineNumToRun; index++ {
					wg.Add(1)
					go worker()
					fmt.Printf("goroutinesRunning: %d\n", runtime.NumGoroutine())
				}
			}
			goroutineStarted = true
		}

	}

	stopGoroutines := func() {
		// 返回正在执行和排队的任务总数
		goroutinesRunning := runtime.NumGoroutine()
		for index := 0; index < goroutinesRunning; index++ {
			ch<-index
		}
		close(ch)
		wg.Wait()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	cpuPercent(1.0)
	for {
		startGoroutines()
		loads := cpuPercent(1)
		if sumFloat64(loads) >= float64(runtime.NumCPU()%50) {
			fmt.Println(sumFloat64(loads))
			stopGoroutines()
			time.Sleep(time.Duration(5) * time.Second)
		}
		select {
		case _ = <-sc:
			stopGoroutines()
			return
		default:
		}
	}

}

func main() {
	sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
	startLoadUpdate()
}
