package profiler

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type Config struct {
	CpuProfilePath string
	MemProfilePath string
	CollectTime    int64
}

func Collect(cfg Config) error {
	if cfg.CollectTime <= 0 {
		return nil
	}

	fcpu, err := os.Create(cfg.CpuProfilePath)
	if err != nil {
		return fmt.Errorf("could not create CPU profiles file: %w", err)
	}
	defer fcpu.Close()
	if err := pprof.StartCPUProfile(fcpu); err != nil {
		return fmt.Errorf("could not start CPU profiling: %w", err)
	}
	defer pprof.StopCPUProfile()
	time.Sleep(time.Duration(cfg.CollectTime) * time.Second)

	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(cfg.MemProfilePath)
	if err != nil {
		return fmt.Errorf("could not create memory profiles file: %w", err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		return fmt.Errorf("could not write memory profiles to file: %w", err)
	}

	return nil
}
