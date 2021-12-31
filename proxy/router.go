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

func New(config *Config, routable Routable) *Server {
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

	return &Server{
		server: &fasthttp.Server{
			Handler:     router.HandleRequest,
			Concurrency: config.ConcurrencyLimit,
		},
		config: config,
	}
}

type Server struct {
	config *Config
	server *fasthttp.Server
}

func (s *Server) Serve() {
	logrus.Infof("starting proxy server on :%s port...", s.config.Port)
	logrus.Fatal(s.server.ListenAndServe(fmt.Sprintf("%s:%s", s.config.Bindaddr, s.config.Port)))
}
