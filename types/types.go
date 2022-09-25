package types

import "fmt"

type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

type StdOutLogger struct{}

func (l StdOutLogger) Debug(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Info(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Warn(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Error(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
}
