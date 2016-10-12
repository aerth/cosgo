// +build debug

// Build with "make debug" to get access to the /debugging url
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http/pprof"
	_ "net/http/pprof"
	"time"
)

func init() {
	go func() {
		time.Sleep(time.Second * 1)
		fmt.Println("DEBUG BUILD")
		fmt.Println("DEBUG BUILD")
		fmt.Println("DEBUG BUILD")
		fmt.Println("DEBUG BUILD")
	}()

	log.SetFlags(log.LstdFlags + log.Llongfile)
	fmt.Println("DEBUG BUILD")
	flag.Usage = func() {
		fmt.Println("DEBUG BUILD")
	}

}

func (c *Cosgo) debug() {
	fmt.Println("Registering /debug/pprof/ URL handlers")
	c.r.HandleFunc("/debug/pprof/", pprof.Index)
	c.r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	c.r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	c.r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	c.r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	c.r.Handle("/debug/pprof/block", pprof.Handler("block"))
	c.r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	c.r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	c.r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	c.r.Handle("/profilez", pprof.Handler("profilez"))

	return
}
