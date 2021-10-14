/*
 * Copyright (c) 2021, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package resources

import (
	"os"
)

const (
	memCgroup2LimitPath = "/sys/fs/cgroup/memory.max"
	memCgroup2UsagePath = "/sys/fs/cgroup/memory.current"
)

type cgroup2RAM struct {
	usageFile *os.File
	limit     uint64
}

func newCgroup2RAM() (*cgroup2RAM, error) {
	var (
		err       error
		limitFile *os.File
		c         = &cgroup2RAM{}
	)

	if limitFile, err = os.Open(memCgroup2LimitPath); err != nil {
		return nil, err
	}
	defer limitFile.Close()
	if c.limit, err = readUint64FromFile(limitFile); err != nil {
		return nil, err
	}

	if c.usageFile, err = os.Open(memCgroup2UsagePath); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cgroup2RAM) tick() (*RAM, error) {
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

func (c *cgroup2RAM) getRAMLimitMegabytes() uint64 {
	return c.limit >> 20
}
