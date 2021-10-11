package resources

import (
	"os"
)

const (
	memCgroupLimitPath = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	memCgroupUsagePath = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
)

type cgroupRAM struct {
	usageFile *os.File
	limit     uint64
}

func newCgroupRAM() (*cgroupRAM, error) {
	var (
		err       error
		limitFile *os.File
		c         = &cgroupRAM{}
	)

	if limitFile, err = os.Open(memCgroupLimitPath); err != nil {
		return nil, err
	}
	defer limitFile.Close()
	if c.limit, err = readUint64FromFile(limitFile); err != nil {
		return nil, err
	}

	if c.usageFile, err = os.Open(memCgroupUsagePath); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cgroupRAM) tick() (*RAM, error) {
	var (
		err  error
		used uint64
	)

	if used, err = readUint64FromFile(c.usageFile); err != nil {
		return nil, err
	}

	return &RAM{
		Usage: float64(used) / float64(c.limit),
		Total: c.limit >> 10,
		Used:  used >> 10,
		Free:  (c.limit - used) >> 10,
	}, nil
}

func (c *cgroupRAM) getRAMLimitMegabytes() uint64 {
	return c.limit >> 20
}
