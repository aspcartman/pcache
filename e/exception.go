package e

import "fmt"

/*
	Package `e` provides a simple high-level way of error handling that may
	be used in higher-order packages, where classic go-way error handling turns into
	`if err != nil { return nil, err }` mantra. It's not intended to be used in a library code.
 */

type Exception struct {
	Error       error
	Description []interface{}
}

func (ex *Exception) Info() string {
	return fmt.Sprint(ex.Description...)
}

func Catch(handler func(e *Exception))  {
	if r := recover(); r != nil {
		handle(r, handler)
	}
}

func OnError(handler func(e *Exception)) {
	if r := recover(); r != nil {
		handle(r, handler)
		panic(r)
	}
}

func Throw(description ... interface{}) {
	exception := formException(description)
	ExecHooks(exception)
	panic(exception)
}

func Must(err error) {
	if err != nil {
		Throw(err)
	}
}


func handle(r interface{}, handler func(e *Exception)) {
	var exception *Exception
	switch e := r.(type) {
	case *Exception:
		exception = e
	case error:
		exception = &Exception{e, nil}
	default:
		exception = &Exception{nil, []interface{}{e}}
	}

	handler(exception)
}

func formException(description []interface{}) *Exception {
	if len(description) == 0 {
		return &Exception{nil, []interface{}{"unknown reason"}}
	}

	// find the first err
	var err error
	for i, r := range description {
		if re, ok := r.(error); ok {
			err = re
			description = description[:i+copy(description[i:], description[i+1:])] // delete err from exception description
			break
		}
	}

	return &Exception{err, description}
}
