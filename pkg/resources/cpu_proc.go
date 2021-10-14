/*
 * Copyright (c) 2021, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package resources

import (
	"bufio"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cpuProcStatPath = "/proc/stat"
)

type procCPU struct {
	prevIdleTime, prevTotalTime uint64
	floatingAvgSeconds          int
	accumulatedCPUUsage         float64
	lastError                   error
	statFile                    *os.File
	once                        sync.Once
}

func newProcCPU() (*procCPU, error) {
	var (
		err error
		p   = &procCPU{
			floatingAvgSeconds:  5,
			accumulatedCPUUsage: 0.0,
		}
	)

	if p.statFile, err = os.Open(cpuProcStatPath); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *procCPU) setFloatingAvgSeconds(seconds int) {
	c.floatingAvgSeconds = seconds
}

func (p *procCPU) tick() (*CPU, error) {
	go p.cpuUsageRunner()
	return &CPU{Usage: p.accumulatedCPUUsage}, p.lastError
}

func (p *procCPU) getCPUMillicores() uint64 {
	return uint64(runtime.NumCPU() * 1000)
}

func (p *procCPU) cpuUsageRunner() {
	p.once.Do(func() {
		var (
			cpuUsages = make([]float64, 10*p.floatingAvgSeconds)
			usage, l  float64
			index, i  uint64
			acc       float64
		)

		for {
			if usage, p.lastError = p.partialCPUUsage(); p.lastError != nil {
				continue
			}

			cpuUsages[index%uint64(len(cpuUsages))] = usage
			acc = 0.0
			for i = 0; i < uint64(len(cpuUsages)); i++ {
				if i <= index {
					acc += cpuUsages[i]
				}
			}

			l = float64(len(cpuUsages))
			if index < uint64(len(cpuUsages)) {
				l = float64(index)
			}
			p.accumulatedCPUUsage = acc / l

			index++
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (p *procCPU) partialCPUUsage() (float64, error) {
	var (
		cpuUsage                                              float64
		s, firstLine                                          string
		split                                                 []string
		scanner                                               *bufio.Scanner
		totalTime, idleTime, deltaIdleTime, deltaTotalTime, u uint64
		err                                                   error
	)

	if _, err = p.statFile.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	scanner = bufio.NewScanner(p.statFile)
	scanner.Scan()
	firstLine = scanner.Text()[5:]

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	split = strings.Fields(firstLine)
	idleTime, _ = strconv.ParseUint(split[3], 10, 64)
	totalTime = uint64(0)
	for _, s = range split {
		u, _ = strconv.ParseUint(s, 10, 64)
		totalTime += u
	}
	if p.prevIdleTime > 0 && p.prevTotalTime > 0 {
		deltaIdleTime = idleTime - p.prevIdleTime
		deltaTotalTime = totalTime - p.prevTotalTime
		cpuUsage = (1.0 - float64(deltaIdleTime)/float64(deltaTotalTime))
	}
	p.prevIdleTime = idleTime
	p.prevTotalTime = totalTime
	return cpuUsage, nil

}
