/*
 * Copyright (c) 2021, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package resources

import (
	"os"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

const (
	memProcMeminfoPath = "/proc/meminfo"
)

type procRAM struct{}

func newProcRAM() (*procRAM, error) {
	if _, err := os.Stat(memProcMeminfoPath); os.IsNotExist(err) {
		return nil, err
	}
	return &procRAM{}, nil
}

func (p *procRAM) tick() (*RAM, error) {
	var (
		stat *linuxproc.MemInfo
		err  error
	)

	if stat, err = linuxproc.ReadMemInfo(memProcMeminfoPath); err != nil {
		return nil, err
	}

	return &RAM{
		Usage: float64(stat.MemTotal-stat.MemAvailable) / float64(stat.MemTotal),
		Total: stat.MemTotal,
		Used:  stat.MemTotal - stat.MemAvailable,
		Free:  stat.MemAvailable,
	}, nil
}

func (p *procRAM) getRAMLimitMegabytes() uint64 {
	var (
		stat *linuxproc.MemInfo
		err  error
	)
	if stat, err = linuxproc.ReadMemInfo(memProcMeminfoPath); err != nil {
		return 0
	}
	return stat.MemTotal >> 10
}
