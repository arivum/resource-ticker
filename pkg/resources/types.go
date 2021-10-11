package resources

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func SetLogger(log *logrus.Logger) {
	logger = log
}

type ResourceTicker struct {
	ram      genericRAM
	cpu      genericCPU
	tickOnce sync.Once
	Events   chan Resources
	Errors   chan error
}

type Resources struct {
	RAM *RAM
	CPU *CPU
}

type genericRAM interface {
	tick() (*RAM, error)
	getRAMLimitMegabytes() uint64
}

type genericCPU interface {
	tick() (*CPU, error)
	getCPUMillicores() uint64
	setFloatingAvgSeconds(seconds int)
}

type CPU struct {
	Usage float64
}

type RAM struct {
	Usage float64
	Total uint64
	Used  uint64
	Free  uint64
}
