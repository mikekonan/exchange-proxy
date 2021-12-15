package main

import (
	"fmt"
	"os"
	"time"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/jaffee/commandeer"
	"github.com/mikekonan/freqtradeProxy/proxy"
	"github.com/mikekonan/freqtradeProxy/proxy/kucoin"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/sirupsen/logrus"
)

var version = "dev"

type app struct {
	Verbose         int           `help:"verbose level: 0 - info, 1 - debug, 2 - trace"`
	CacheSize       int           `help:"amount of candles to cache"`
	TTLCacheTimeout time.Duration `help:"ttl of blobs of cached data"`

	ProxyConfig  proxy.Config  `flag:"!embed"`
	KucoinConfig kucoin.Config `flag:"!embed"`
}

func newApp() *app {
	return &app{
		Verbose:         0,
		CacheSize:       1000,
		TTLCacheTimeout: time.Minute * 10,
		KucoinConfig: kucoin.Config{
			TopicsPerWs: 200,
			RequestURL:  "https://openapi-v2.kucoin.com",
		},
		ProxyConfig: proxy.Config{
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

	logrus.Infof("freqtrade-proxy version - %s", version)

	if a.Verbose > 2 {
		return fmt.Errorf("wrong verbose level '%d'", a.Verbose)
	}

	a.configure()

	if err := a.ProxyConfig.Validate(); err != nil {
		return err
	}

	if err := a.KucoinConfig.Validate(); err != nil {
		return err
	}

	proxySrv := proxy.New(&a.ProxyConfig, kucoin.New(
		store.NewStore(a.CacheSize),
		store.NewTTLCache(a.TTLCacheTimeout),
		&a.KucoinConfig),
	)
	proxySrv.Serve()

	return nil
}

func main() {
	app := newApp()

	if err := commandeer.Run(app); err != nil {
		logrus.Fatal(err)
	}
}
