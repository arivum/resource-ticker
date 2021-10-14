/*
 * Copyright (c) 2021, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package resources

import (
	"time"

	"github.com/hashicorp/go-multierror"
)

type Option func(*ResourceTicker)

func WithCPUFloatingAvg(seconds int) Option {
	return func(r *ResourceTicker) { r.cpu.setFloatingAvgSeconds(seconds) }
}

func NewResourceTicker(options ...Option) (*ResourceTicker, error) {
	var (
		ticker = &ResourceTicker{
			Events: make(chan Resources),
		}
		err    error
		mErr   multierror.Error
		option Option
	)

	if ticker.cpu, err = newCgroup2CPU(); err != nil {
		mErr.Errors = append(mErr.Errors, err)
		if ticker.cpu, err = newCgroupCPU(); err != nil {
			mErr.Errors = append(mErr.Errors, err)
			if ticker.cpu, err = newProcCPU(); err != nil {
				mErr.Errors = append(mErr.Errors, err)
				return nil, &mErr
			} else {
				logger.Debug("calculating CPU usage from proc fs")
			}
		} else {
			logger.Debug("calculating CPU usage from cgroups")
		}
	} else {
		logger.Debug("calculating CPU usage from cgroups2")
	}

	mErr = multierror.Error{}
	if ticker.ram, err = newCgroup2RAM(); err != nil {
		mErr.Errors = append(mErr.Errors, err)
		if ticker.ram, err = newCgroupRAM(); err != nil {
			mErr.Errors = append(mErr.Errors, err)
			if ticker.ram, err = newProcRAM(); err != nil {
				mErr.Errors = append(mErr.Errors, err)
				return nil, &mErr
			} else {
				logger.Debug("gathering memory information from proc fs")
			}
		} else {
			logger.Debug("gathering memory information from cgroups")
		}
	} else {
		logger.Debug("gathering memory information from cgroups2")
	}

	for _, option = range options {
		option(ticker)
	}

	return ticker, nil
}

func (r *ResourceTicker) Run() (chan Resources, chan error) {
	go r.tickOnce.Do(r.tickerFunc)
	return r.Events, r.Errors
}

func (r *ResourceTicker) GetCPUMillicores() uint64 {
	return r.cpu.getCPUMillicores()
}

func (r *ResourceTicker) GetRAMLimitMegabytes() uint64 {
	return r.ram.getRAMLimitMegabytes()
}

func (r *ResourceTicker) tickerFunc() {
	var (
		cpu *CPU
		ram *RAM
		err error
	)
	for {
		if cpu, err = r.cpu.tick(); err == errNotEnoughDataPoints {
			continue
		} else if err != nil {
			r.Errors <- err
		}
		if ram, err = r.ram.tick(); err != nil {
			r.Errors <- err
		}
		r.Events <- Resources{
			RAM: ram,
			CPU: cpu,
		}
		time.Sleep(1 * time.Second)
	}
}
