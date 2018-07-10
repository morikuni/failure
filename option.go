package failure

// Option represents an optional parameter of failure.
type Option interface {
	ApplyTo(*Failure)
}

// OptionFunc represents an option with function.
type OptionFunc func(*Failure)

// ApplyTo implements the interface Option.
func (of OptionFunc) ApplyTo(f *Failure) {
	of(f)
}

// Message adds a message to a failure.
func Message(msg string) Option {
	return OptionFunc(func(f *Failure) {
		f.Message = msg
	})
}

// Info is key-value data.
type Info map[string]interface{}

// ApplyTo implements the interface Option.
func (i Info) ApplyTo(f *Failure) {
	f.Info = i
}
