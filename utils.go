package main

import (
	"bufio"
	"fmt"
	"os"
)

func getenv[T any](key string, fallback T, parse func(string) (T, error)) T {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := parse(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func check_err(e error) {
	if e != nil {
		panic(e)
	}
}

func exit_with_one(message string) {
	if debug {
		panic(message)
	} else {
		fmt.Println(message)
		os.Exit(1)
	}
}

func assert_runtime(cond bool, message string) {
	if !cond {
		fmt.Println("Runtime Assertion Error")
		panic(message)
	}
}

func prompt_for_debug() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n-> Press Enter")
	_, _, err := reader.ReadRune()
	check_err(err)
	fmt.Println()
}
