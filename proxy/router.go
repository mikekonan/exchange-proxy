package proxy

import (
	"fmt"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Routable interface {
	Routes() map[string]struct {
		Method  string
		Handler func(c *routing.Context) error
	}

	Name() string
}

func New(config *Config, routable Routable) *Router {
	router := routing.New()

	for k, v := range routable.Routes() {
		path := fmt.Sprintf("/%s/%s", routable.Name(), k)
		logrus.Infof("applying route '%s' of method '%s'", path, v.Method)

		if v.Method == "<ANY>" {
			router.Any(path, v.Handler)
			continue
		}

		router.To(v.Method, path, v.Handler)
	}

	return &Router{
		router: router,
		config: config,
	}
}

type Router struct {
	config *Config
	router *routing.Router
}

func (r *Router) Serve() {
	logrus.Infof("starting proxy server on :%s port...", r.config.Port)
	panic(fasthttp.ListenAndServe(fmt.Sprintf("%s:%s", r.config.Bindaddr, r.config.Port), r.router.HandleRequest))
}
