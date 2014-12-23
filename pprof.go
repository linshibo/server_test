package main

import (
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	//"net/http/pprof"
	"sanguo/base/log"
	"strings"
	"sync"
	"time"
)

var cpuProfile *os.File
var cpuProfileLock sync.Mutex

func closeCPUProf() {
	select {
	case <-time.After(time.Second * 5):
		LookUp("stop cpuprof")
	}
}

func LookUp(op string) {
	timestamp := strings.Replace(time.Now().String()[0:19], " ", "-", 1)
	switch op {
	case "lookup heap":
		if f, err := os.Create("heap_prof" + "." + timestamp); err != nil {
			log.Debug("heap profile failed:", err)
		} else {
			p := pprof.Lookup("heap")
			p.WriteTo(f, 2)
		}
	case "lookup goroutine":
		if f, err := os.Create("goroutine_prof" + "." + timestamp); err != nil {
			log.Debug("goroutine profile failed:", err)
		} else {
			p := pprof.Lookup("goroutine")
			p.WriteTo(f, 2)
		}

	case "lookup threadcreate":
		if f, err := os.Create("threadcreate_prof" + "." + timestamp); err != nil {
			log.Debug("threadcreate profile failed:", err)
		} else {
			p := pprof.Lookup("threadcreate")
			p.WriteTo(f, 2)
		}
	case "lookup block":
		if f, err := os.Create("block_prof" + "." + timestamp); err != nil {
			log.Debug("block profile failed:", err)
		} else {
			p := pprof.Lookup("block")
			p.WriteTo(f, 2)
		}
	case "start cpuprof":
		cpuProfileLock.Lock()
		defer cpuProfileLock.Unlock()
		if cpuProfile == nil {
			if f, err := os.Create("cpu_prof" + "." + timestamp); err != nil {
				log.Debug("start cpu profile failed: %v", err)
			} else {
				log.Debug("start cpu profile")
				pprof.StartCPUProfile(f)
				cpuProfile = f
			}
		}
		go closeCPUProf()
	case "stop cpuprof":
		cpuProfileLock.Lock()
		defer cpuProfileLock.Unlock()
		if cpuProfile != nil {
			pprof.StopCPUProfile()
			cpuProfile.Close()
			cpuProfile = nil
			log.Debug("stop cpu profile")
		}
	case "lookup memprof":
		if f, err := os.Create("mem_prof" + "." + timestamp); err != nil {
			log.Debug("record memory profile failed: %v", err)
		} else {
			runtime.GC()
			pprof.WriteHeapProfile(f)
			f.Close()
			log.Debug("record memory profile")
		}
	case "read memstats":
		m := new(runtime.MemStats)
		runtime.ReadMemStats(m)
		log.Debugf("memstats %+v", *m)
	case "free mem":
		debug.FreeOSMemory()
	case "heap dump":
		if f, err := os.Create("heapdump_prof" + "." + timestamp); err != nil {
			log.Debug("record memory profile failed: %v", err)
		} else {
			debug.WriteHeapDump(f.Fd())
			f.Close()
			log.Debug("heapdump profile")
		}
	case "run gc":
		runtime.GC()
	}
}
