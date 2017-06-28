package diff

type config struct {
	sliceFn diffFn
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
