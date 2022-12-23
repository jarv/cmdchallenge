package main

import (
	"flag"
	"time"
)

const defaultTimeout int = 1000

func main() {
	timeout := flag.Int("timeout", defaultTimeout, "how many seconds to sleep")
	flag.Parse()

	for i := 0; ; i++ {
		if timeout != nil && i >= *timeout {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
