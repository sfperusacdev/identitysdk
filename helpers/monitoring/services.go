package monitoring

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type MetricsService struct {
	start time.Time
	proc  *process.Process
}

func NewMetricsService() *MetricsService {
	p, _ := process.NewProcess(int32(os.Getpid()))
	return &MetricsService{start: time.Now(), proc: p}
}

func (s *MetricsService) Collect() Response {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	vm, _ := mem.VirtualMemory()
	sm, _ := mem.SwapMemory()
	ld, _ := load.Avg()
	hu, _ := host.Uptime()
	ds, _ := disk.Usage("/")

	threads, _ := s.proc.NumThreads()
	fds, _ := s.proc.NumFDs()

	return Response{
		Status: "ok",
		Time:   time.Now().UTC(),
		Uptime: time.Since(s.start).Seconds(),
		Process: Process{
			PID:            os.Getpid(),
			Goroutines:     runtime.NumGoroutine(),
			Threads:        threads,
			OpenFDs:        fds,
			HeapAlloc:      ms.HeapAlloc,
			TotalAlloc:     ms.TotalAlloc,
			GCCycles:       ms.NumGC,
			GCPauseTotalNs: ms.PauseTotalNs,
			NextGCBytes:    ms.NextGC,
		},
		Runtime: Runtime{
			NumCPU: runtime.NumCPU(),
		},
		Host: Host{
			Load1:  ld.Load1,
			Load5:  ld.Load5,
			Load15: ld.Load15,
			Uptime: float64(hu),
			Memory: Memory{
				UsedBytes:     vm.Used,
				TotalBytes:    vm.Total,
				FreeBytes:     vm.Available,
				UsedPct:       vm.UsedPercent,
				SwapUsedBytes: sm.Used,
				SwapUsedPct:   sm.UsedPercent,
			},
			Disk: Disk{
				UsedBytes:  ds.Used,
				TotalBytes: ds.Total,
				FreeBytes:  ds.Free,
				UsedPct:    ds.UsedPercent,
				InodesPct:  ds.InodesUsedPercent,
			},
		},
	}
}
