package main

import (
	_ "net/http/pprof"
	HTTPTerminal "shopee/entry_task/HTTP"
	TCPTerminal "shopee/entry_task/TCP"
	"sync"
)

func main() {
	wg := sync.WaitGroup{} // number of working goroutines
	wg.Add(2)

	// HTTP
	go func() {
		HTTPTerminal.Run()
		wg.Done()
	}()

	// TCP
	go func() {
		TCPTerminal.Run()
		wg.Done()
	}()

	wg.Wait()
}

// go run ./main.go & wrk -t200 -c1000 -s ./test/mulTest.lua -d30s http://127.0.0.1:5500 & go tool pprof http://localhost:5500/debug/pprof/profile
