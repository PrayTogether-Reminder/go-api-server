package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var env string
	flag.StringVar(&env, "env", "local", "Environment (local|dev|prod)")
	flag.Parse()

	fmt.Printf("Starting server in %s environment...\n", env)
	os.Exit(0)
}
