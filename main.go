package main

import (
	"fmt"
	"os"
	"time"

	logrusStack "github.com/Gurpartap/logrus-stack"
	"github.com/jaffee/commandeer"
	"github.com/mikekonan/freqtradeProxy/proxy"
	"github.com/mikekonan/freqtradeProxy/proxy/kucoin"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var version = "dev"

type app struct {
	Verbose         int           `help:"verbose level: 0 - info, 1 - debug, 2 - trace"`
	CacheSize       int           `help:"amount of candles to cache"`
	TTLCacheTimeout time.Duration `help:"ttl of blobs of cached data"`
	ClientTimeout   time.Duration `help:"client timeout"`

	ProxyConfig  proxy.Config  `flag:"!embed"`
	KucoinConfig kucoin.Config `flag:"!embed"`
}

func newApp() *app {
	return &app{
		Verbose:         0,
		CacheSize:       1000,
		TTLCacheTimeout: time.Minute * 10,
		ClientTimeout:   time.Second * 15,
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

func (app *app) configure() {
	switch app.Verbose {
	case 0:
		logrus.SetLevel(logrus.InfoLevel)
	case 1:
		logrus.SetLevel(logrus.DebugLevel)
	case 2:
		logrus.SetLevel(logrus.TraceLevel)
	}
}

func (app *app) Run() error {
	logrus.SetOutput(os.Stdout)
	logrus.AddHook(logrusStack.StandardHook())

	logrus.Infof("starting freqtrade-proxy: version - '%s'... ", version)

	if app.Verbose > 2 {
		return fmt.Errorf("wrong verbose level '%d'", app.Verbose)
	}

	app.configure()

	if err := app.ProxyConfig.Validate(); err != nil {
		return err
	}

	if err := app.KucoinConfig.Validate(); err != nil {
		return err
	}

	client := &proxy.Client{
		Client: fasthttp.Client{
			ReadTimeout:  app.ClientTimeout,
			WriteTimeout: app.ClientTimeout,
		},
	}

	proxySrv := proxy.New(&app.ProxyConfig,
		kucoin.New(
			store.NewStore(app.CacheSize),
			store.NewTTLCache(app.TTLCacheTimeout),
			client,
			&app.KucoinConfig,
		),
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
