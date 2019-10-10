package main

import (
	"flag"
	"github.com/thebagchi/heimdall-proxy/pkg/proxy"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	var (
		address   = flag.String("address", "", "bind address")
		forward   = flag.String("forward", "", "forward address")
		duplicate = flag.String("duplicate", "", "duplicate address")
	)

	flag.Parse()

	if strings.Compare(*address, "") == 0 {
		log.Fatalln("Cannot start proxy server, bind address is not provided...")
	}

	if strings.Compare(*forward, "") == 0 {
		log.Fatalln("Cannot start proxy server, forward address is not provided...")
	}

	var (
		signals = make(chan os.Signal, 1)
		done    = make(chan bool, 1)
	)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		done <- true
	}()

	go func() {
		err := proxy.StartServer(*address, *forward, *duplicate)
		if nil != err {
			log.Fatalln("Failed starting proxy server, cause: ", err.Error())
		}
	}()

	<-done
}
