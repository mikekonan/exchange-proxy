package proxy

import (
	"fmt"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const AnyHTTPMethod = "<ANY>"

type Routable interface {
	Routes() []struct {
		Path    string
		Method  string
		Handler func(c *routing.Context) error
	}

	Name() string
}

func New(config *Config, routable Routable) *Router {
	router := routing.New()

	for _, route := range routable.Routes() {
		path := fmt.Sprintf("/%s/%s", routable.Name(), route.Path)
		logrus.Infof("applying route '%s' of method '%s'", path, route.Method)

		if route.Method == AnyHTTPMethod {
			router.Any(path, route.Handler)
			continue
		}

		router.To(route.Method, path, route.Handler)
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
