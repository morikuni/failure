package failure

// Option represents an optional parameter of failure.
type Option func(*Failure)

// WithMessage adds a message to a failure.
func WithMessage(msg string) Option {
	return func(f *Failure) {
		f.Message = msg
	}
}

// WithInfo adds info to a failure.
func WithInfo(info Info) Option {
	return func(f *Failure) {
		f.Info = info
	}
}
