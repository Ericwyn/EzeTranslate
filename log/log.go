package log

import "fmt"

func E(out ...interface{}) {
	fmt.Println("[LOG]", fmt.Sprint(out...))
}

func D(out ...interface{}) {
	fmt.Println("[LOG]", fmt.Sprint(out...))
}
