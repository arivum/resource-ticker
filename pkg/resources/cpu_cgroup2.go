package resources

import (
	"bufio"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	cpuCgroup2MaxPath   = "/sys/fs/cgroup/cpu.max"
	cpuCgroup2UsagePath = "/sys/fs/cgroup/cpu.stat"
)

type cgroup2CPU struct {
	floatingAvgSeconds int
	cpuUsages          []float64
	cores              float64
	usageFile          *os.File
	lastTime           int64
	lastUsage          uint64
	index              uint64
}

func newCgroup2CPU() (*cgroup2CPU, error) {
	var (
		floatingAvgSeconds = 5
		err                error
		maxFile            *os.File
		quota, period      uint64
		c                  = &cgroup2CPU{
			floatingAvgSeconds: floatingAvgSeconds,
			cpuUsages:          make([]float64, floatingAvgSeconds),
			lastTime:           0,
			lastUsage:          0,
		}
		line      []byte
		splitLine []string
	)

	if maxFile, err = os.Open(cpuCgroup2MaxPath); err != nil {
		return nil, err
	}
	defer maxFile.Close()

	if line, _, err = bufio.NewReader(maxFile).ReadLine(); err != nil {
		return nil, err
	}

	splitLine = strings.Split(string(line), " ")
	if len(splitLine) < 2 {
		return nil, errNoCgroup2CPULimit
	}
	if quota, err = strconv.ParseUint(strings.TrimSpace(splitLine[0]), 10, 64); err != nil {
		return nil, err
	}
	if period, err = strconv.ParseUint(strings.TrimSpace(splitLine[1]), 10, 64); err != nil {
		return nil, err
	}

	c.cores = float64(quota) / float64(period)

	if c.usageFile, err = os.Open(cpuCgroup2UsagePath); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *cgroup2CPU) setFloatingAvgSeconds(seconds int) {
	if seconds > 1 {
		c.floatingAvgSeconds = seconds
		c.cpuUsages = make([]float64, c.floatingAvgSeconds)
	}
}

func (c *cgroup2CPU) getCurrentNs() (uint64, error) {
	var (
		usageReader *bufio.Reader
		err         error
		line        []byte
		splitLine   []string
		usage       uint64
	)

	if _, err = c.usageFile.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	usageReader = bufio.NewReader(c.usageFile)
	if line, _, err = usageReader.ReadLine(); err != nil {
		return 0, err
	}

	if splitLine = strings.Split(string(line), " "); len(splitLine) < 2 {
		return 0, errNoCgroup2CPUStats
	}

	if usage, err = strconv.ParseUint(strings.TrimSpace(splitLine[1]), 10, 64); err != nil {
		return 0, err
	}
	return usage, nil
}

func (c *cgroup2CPU) tick() (*CPU, error) {
	var (
		err      error
		newTime  int64
		newUsage uint64
		avgUsage float64
	)

	if c.lastTime == 0 {
		c.lastTime = time.Now().UnixNano()
		if c.lastUsage, err = c.getCurrentNs(); err != nil {
			return nil, err
		}
		return nil, errNotEnoughDataPoints
	} else {
		newTime = time.Now().UnixNano()
		if newUsage, err = c.getCurrentNs(); err != nil {
			return nil, err
		}
		c.cpuUsages[c.index%uint64(len(c.cpuUsages))] = math.Min(float64(newUsage-c.lastUsage)/(float64(newTime-c.lastTime)/1000.0)/c.cores, 1.0)

		c.lastTime = newTime
		c.lastUsage = newUsage
		avgUsage = c.accumulate()

		c.index++

		return &CPU{Usage: avgUsage}, nil
	}
}

func (c *cgroup2CPU) getCPUMillicores() uint64 {
	return uint64(c.cores * 1000.0)
}

func (c *cgroup2CPU) accumulate() float64 {
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
