package main

import (
	"flag"
	"os"

	"github.com/Gurpartap/logrus-stack"
	"github.com/mikekonan/freqtradeProxy/proxy/kucoin"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.AddHook(logrus_stack.StandardHook())

	port := flag.Int("port", 8080, "listen port")
	verbose := flag.Int("verbose", 0, "verbose level. 0 - info [default]. 1 - debug. 2 - trace.")
	flag.Parse()

	switch *verbose {
	case 0:
		logrus.SetLevel(logrus.InfoLevel)
	case 1:
		logrus.SetLevel(logrus.DebugLevel)
	case 2:
		logrus.SetLevel(logrus.TraceLevel)
	}

	s := store.New()
	k := kucoin.New(s)
	k.Start(*port)
}
