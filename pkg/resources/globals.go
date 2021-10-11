package resources

import "errors"

var (
	errNotEnoughDataPoints = errors.New("not enough data points")
	errNoCgroup2CPULimit   = errors.New("it seems like no cgroup2 cpu limits are in place")
	errNoCgroup2CPUStats   = errors.New("it seems like no cgroup2 cpu stats can be read")
)
