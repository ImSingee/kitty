package tl

type option struct {
	inited      bool
	exitOnError bool
}

func defaultOption() option {
	return option{
		inited:      true,
		exitOnError: true,
	}
}

type OptionApplier func(o *option)

func WithExitOnError(exitOnError bool) OptionApplier {
	return func(o *option) {
		o.exitOnError = exitOnError
	}
}
