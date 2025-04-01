// Package profiler simple mem and cpu collector.
package profiler

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Config profiler collector config.
type Config struct {
	CPUProfilePath string // where to store CPU profile
	MemProfilePath string // where to store mem profile
	CollectTime    int64  // how long to collect data
}

// Collect mem and CPU data from profiler.
func Collect(cfg Config) error {
	if cfg.CollectTime <= 0 {
		return nil
	}

	fcpu, err := os.Create(cfg.CPUProfilePath)
	if err != nil {
		return fmt.Errorf("could not create CPU profiles file: %w", err)
	}
	defer fcpu.Close() //nolint:all // its safe
	if err := pprof.StartCPUProfile(fcpu); err != nil {
		return fmt.Errorf("could not start CPU profiling: %w", err)
	}
	defer pprof.StopCPUProfile()
	time.Sleep(time.Duration(cfg.CollectTime) * time.Second)

	fmem, err := os.Create(cfg.MemProfilePath)
	if err != nil {
		return fmt.Errorf("could not create memory profiles file: %w", err)
	}
	defer fmem.Close() //nolint:all // its safe
	runtime.GC()
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		return fmt.Errorf("could not write memory profiles to file: %w", err)
	}

	return nil
}
