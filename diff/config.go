package diff

import "strings"

type config struct {
	sliceFn  diffFn
	sortKeys []string
}

// ConfigOpt is used to pass configuration options to the diff algorithm
type ConfigOpt func(config) config

func defaultConfig() config {
	return config{
		sliceFn: newSlice,
	}
}

// UseSliceMyers configures the Diff function to use Myers' algorithm for slices
func UseSliceMyers() ConfigOpt {
	return func(c config) config {
		c.sliceFn = newMyersSlice
		return c
	}
}

func AddSortingKeys(keys string) ConfigOpt {
	return func(c config) config {
		c.sortKeys = strings.Split(keys, ",")
		return c
	}
}
