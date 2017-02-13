// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package metrics provides general system and process level metrics collection.
package metrics

import (
	"runtime"
	"time"

	"github.com/ethereumproject/go-ethereum/logger"
	"github.com/ethereumproject/go-ethereum/logger/glog"
	"github.com/rcrowley/go-metrics"
)

// NewMeter create a new metrics Meter, either a real one of a NOP stub depending
// on the metrics flag.
func NewMeter(name string) metrics.Meter {
	return metrics.GetOrRegisterMeter(name, metrics.DefaultRegistry)
}

// NewTimer create a new metrics Timer, either a real one of a NOP stub depending
// on the metrics flag.
func NewTimer(name string) metrics.Timer {
	return metrics.GetOrRegisterTimer(name, metrics.DefaultRegistry)
}

// CollectProcessMetrics periodically collects various metrics about the running
// process.
func CollectProcessMetrics(refresh time.Duration) {
	// Create the various data collectors
	memstats := make([]*runtime.MemStats, 2)
	diskstats := make([]*DiskStats, 2)
	for i := 0; i < len(memstats); i++ {
		memstats[i] = new(runtime.MemStats)
		diskstats[i] = new(DiskStats)
	}
	// Define the various metrics to collect
	memAllocs := metrics.GetOrRegisterMeter("system/memory/allocs", metrics.DefaultRegistry)
	memFrees := metrics.GetOrRegisterMeter("system/memory/frees", metrics.DefaultRegistry)
	memInuse := metrics.GetOrRegisterMeter("system/memory/inuse", metrics.DefaultRegistry)
	memPauses := metrics.GetOrRegisterMeter("system/memory/pauses", metrics.DefaultRegistry)

	var diskReads, diskReadBytes, diskWrites, diskWriteBytes metrics.Meter
	if err := ReadDiskStats(diskstats[0]); err == nil {
		diskReads = metrics.GetOrRegisterMeter("system/disk/readcount", metrics.DefaultRegistry)
		diskReadBytes = metrics.GetOrRegisterMeter("system/disk/readdata", metrics.DefaultRegistry)
		diskWrites = metrics.GetOrRegisterMeter("system/disk/writecount", metrics.DefaultRegistry)
		diskWriteBytes = metrics.GetOrRegisterMeter("system/disk/writedata", metrics.DefaultRegistry)
	} else {
		glog.V(logger.Debug).Infof("failed to read disk metrics: %v", err)
	}
	// Iterate loading the different stats and updating the meters
	for i := 1; ; i++ {
		runtime.ReadMemStats(memstats[i%2])
		memAllocs.Mark(int64(memstats[i%2].Mallocs - memstats[(i-1)%2].Mallocs))
		memFrees.Mark(int64(memstats[i%2].Frees - memstats[(i-1)%2].Frees))
		memInuse.Mark(int64(memstats[i%2].Alloc - memstats[(i-1)%2].Alloc))
		memPauses.Mark(int64(memstats[i%2].PauseTotalNs - memstats[(i-1)%2].PauseTotalNs))

		if ReadDiskStats(diskstats[i%2]) == nil {
			diskReads.Mark(int64(diskstats[i%2].ReadCount - diskstats[(i-1)%2].ReadCount))
			diskReadBytes.Mark(int64(diskstats[i%2].ReadBytes - diskstats[(i-1)%2].ReadBytes))
			diskWrites.Mark(int64(diskstats[i%2].WriteCount - diskstats[(i-1)%2].WriteCount))
			diskWriteBytes.Mark(int64(diskstats[i%2].WriteBytes - diskstats[(i-1)%2].WriteBytes))
		}
		time.Sleep(refresh)
	}
}
