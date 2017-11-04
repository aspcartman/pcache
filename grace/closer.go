package grace

// Closer is an accumulates close functions
// for graceful shutdown. It ignores any errors,
// so user is supposed to log them in the passed
// functions
type Closer struct {
	fns []func()
}

func (c *Closer) Add(closeFunc func()) {
	c.fns = append(c.fns, closeFunc)
}

// Close calls all added functions;
// never return error
func (c *Closer) Close() error {
	for _, f := range c.fns {
		f()
	}
	return nil
}
