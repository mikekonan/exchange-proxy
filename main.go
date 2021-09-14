package main

import (
	"os"

	"github.com/Gurpartap/logrus-stack"
	"github.com/mikekonan/freqtradeProxy/proxy/kucoin"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.AddHook(logrus_stack.StandardHook())
}

func main() {
	s := store.New()
	k := kucoin.New(s)
	if err := k.Connect(); err != nil {
		logrus.Fatal(err)
	}

	k.Start()
}
