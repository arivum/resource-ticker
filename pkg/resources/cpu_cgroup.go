/*
 * Copyright (c) 2021, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package resources

import (
	"math"
	"os"
	"time"
)

const (
	cpuCgroupQuotaPath  = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	cpuCgroupPeriodPath = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	cpuCgroupUsagePath  = "/sys/fs/cgroup/cpu/cpuacct.usage"
)

type cgroupCPU struct {
	floatingAvgSeconds int
	cpuUsages          []float64
	cores              float64
	usageFile          *os.File
	lastTime           int64
	lastUsage          uint64
	index              uint64
}

func newCgroupCPU() (*cgroupCPU, error) {
	var (
		floatingAvgSeconds    = 5
		err                   error
		periodFile, quotaFile *os.File
		quota, period         uint64
		c                     = &cgroupCPU{
			floatingAvgSeconds: floatingAvgSeconds,
			cpuUsages:          make([]float64, floatingAvgSeconds),
			lastTime:           0,
			lastUsage:          0,
		}
	)

	if quotaFile, err = os.Open(cpuCgroupQuotaPath); err != nil {
		return nil, err
	}
	defer quotaFile.Close()
	if quota, err = readUint64FromFile(quotaFile); err != nil {
		return nil, err
	}

	if periodFile, err = os.Open(cpuCgroupPeriodPath); err != nil {
		return nil, err
	}
	defer periodFile.Close()
	if period, err = readUint64FromFile(periodFile); err != nil {
		return nil, err
	}

	c.cores = float64(quota) / float64(period)

	if c.usageFile, err = os.Open(cpuCgroupUsagePath); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cgroupCPU) setFloatingAvgSeconds(seconds int) {
	if seconds > 1 {
		c.floatingAvgSeconds = seconds
		c.cpuUsages = make([]float64, c.floatingAvgSeconds)
	}
}

func (c *cgroupCPU) tick() (*CPU, error) {
	var (
		err      error
		newTime  int64
		newUsage uint64
		avgUsage float64
	)

	if c.lastTime == 0 {
		c.lastTime = time.Now().UnixNano()
		if c.lastUsage, err = readUint64FromFile(c.usageFile); err != nil {
			return nil, err
		}
		return nil, errNotEnoughDataPoints
	} else {
		newTime = time.Now().UnixNano()
		if newUsage, err = readUint64FromFile(c.usageFile); err != nil {
			return nil, err
		}
		c.cpuUsages[c.index%uint64(len(c.cpuUsages))] = math.Min(float64(newUsage-c.lastUsage)/float64(newTime-c.lastTime)/c.cores, 1.0)

		c.lastTime = newTime
		c.lastUsage = newUsage
		avgUsage = c.accumulate()

		c.index++

		return &CPU{Usage: avgUsage}, nil
	}
}

func (c *cgroupCPU) getCPUMillicores() uint64 {
	return uint64(c.cores * 1000.0)
}

func (c *cgroupCPU) accumulate() float64 {
	var (
		l   = float64(len(c.cpuUsages))
		acc = 0.0
		i   = uint64(0)
	)

	for i = uint64(0); i < uint64(len(c.cpuUsages)); i++ {
		if i <= c.index {
			acc += c.cpuUsages[i]
		}
	}

	if c.index+1 < uint64(len(c.cpuUsages)) {
		l = float64(c.index + 1)
	}

	return acc / l
}
