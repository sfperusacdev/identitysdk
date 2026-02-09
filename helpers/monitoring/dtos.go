package monitoring

import "time"

type Response struct {
	Status  string    `json:"status"`
	Time    time.Time `json:"timestamp"`
	Uptime  float64   `json:"uptime_seconds"`
	Process Process   `json:"process"`
	Runtime Runtime   `json:"runtime"`
	Host    Host      `json:"host"`
}

type Process struct {
	PID            int    `json:"pid"`
	Goroutines     int    `json:"goroutines"`
	Threads        int32  `json:"threads"`
	OpenFDs        int32  `json:"open_fds"`
	HeapAlloc      uint64 `json:"heap_alloc_bytes"`
	TotalAlloc     uint64 `json:"total_alloc_bytes"`
	GCCycles       uint32 `json:"gc_cycles"`
	GCPauseTotalNs uint64 `json:"gc_pause_total_ns"`
	NextGCBytes    uint64 `json:"next_gc_bytes"`
}

type Runtime struct {
	NumCPU int `json:"num_cpu"`
}

type Host struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
	Uptime float64 `json:"uptime_seconds"`
	Memory Memory  `json:"memory"`
	Disk   Disk    `json:"disk"`
}

type Memory struct {
	UsedBytes     uint64  `json:"used_bytes"`
	TotalBytes    uint64  `json:"total_bytes"`
	FreeBytes     uint64  `json:"free_bytes"`
	UsedPct       float64 `json:"used_percent"`
	SwapUsedBytes uint64  `json:"swap_used_bytes"`
	SwapUsedPct   float64 `json:"swap_used_percent"`
}

type Disk struct {
	UsedBytes  uint64  `json:"used_bytes"`
	TotalBytes uint64  `json:"total_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	UsedPct    float64 `json:"used_percent"`
	InodesPct  float64 `json:"inodes_used_percent"`
}
