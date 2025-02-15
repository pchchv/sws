package main

import "context"

type MockCommand struct {
	runFunc      func(context.Context) error
	helpFunc     func() string
	describeFunc func() string
}
