package log

import "fmt"

type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

type StdOutLogger struct {
	TransactionID string
	// LogLvl // @todo support different log lvls
}

func (l StdOutLogger) Debug(s string, a ...interface{}) {
	if l.TransactionID != "" {
		fmt.Printf("[trxid: %v][lvl: debug] %v", l.TransactionID, fmt.Sprintf(s, a...))
		return
	}
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Info(s string, a ...interface{}) {
	if l.TransactionID != "" {
		fmt.Printf("[trxid: %v][lvl: info] %v", l.TransactionID, fmt.Sprintf(s, a...))
		return
	}
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Warn(s string, a ...interface{}) {
	if l.TransactionID != "" {
		fmt.Printf("[trxid: %v][lvl: warn] %v", l.TransactionID, fmt.Sprintf(s, a...))
		return
	}
	fmt.Printf(s, a...)
	fmt.Println()
}

func (l StdOutLogger) Error(s string, a ...interface{}) {
	if l.TransactionID != "" {
		fmt.Printf("[trxid: %v][lvl: error] %v", l.TransactionID, fmt.Sprintf(s, a...))
		return
	}
	fmt.Printf(s, a...)
	fmt.Println()
}
