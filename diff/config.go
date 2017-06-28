package diff

type config struct {
	sliceFn diffFn
}

type ConfigOpt func(config) config

func defaultConfig() config {
	return config{
		sliceFn: newSlice,
	}
}

func UseSliceMyers() ConfigOpt {
	return func(c config) config {
		c.sliceFn = newMyersSlice
		return c
	}
}
