package e

var hooks = []func(ex *Exception){}

/*
	Hooks are called every time exception is thrown.
	It is allowed to modify exception.
	Hooks registration suppose to happen at application start only,
	the failure to do so may lead to races and panics.
 */

//go:norace
func RegisterPostHook(f func(ex *Exception)) {
	hooks = append(hooks, f)
}

//go:norace
func ExecHooks(ex *Exception) {
	for _, h := range hooks {
		h(ex)
	}
}
