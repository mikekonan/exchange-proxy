package main

import (
	"fmt"
	"os"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/jaffee/commandeer"
	"github.com/mikekonan/freqtradeProxy/proxy/kucoin"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/sirupsen/logrus"
)

type app struct {
	Config  kucoin.Config `flag:"!embed"`
	Verbose int           `help:"verbose level: 0 - info, 1 - debug, 2 - trace"`
}

func newApp() *app {
	return &app{
		Verbose: 0,
		Config: kucoin.Config{
			Port:     "8080",
			Bindaddr: "0.0.0.0",
		},
	}
}

func (m *app) configure() {
	switch m.Verbose {
	case 0:
		logrus.SetLevel(logrus.InfoLevel)
	case 1:
		logrus.SetLevel(logrus.DebugLevel)
	case 2:
		logrus.SetLevel(logrus.TraceLevel)
	}
}

func (a *app) Run() error {
	logrus.SetOutput(os.Stdout)
	logrus.AddHook(logrus_stack.StandardHook())

	if a.Verbose > 2 {
		return fmt.Errorf("wrong verbose level '%d'", a.Verbose)
	}

	a.configure()

	if err := a.Config.Validate(); err != nil {
		return err
	}

	k := kucoin.New(store.New(), a.Config)
	k.Start()

	return nil
}

func main() {
	app := newApp()

	if err := commandeer.Run(app); err != nil {
		logrus.Fatal(err)
	}
}
