package proxy

import (
	"bytes"
	"net/http"

	"github.com/mikekonan/freqtradeProxy/store"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	contentEncodingHeaderBytes = []byte("Content-Encoding")
	gzipHeaderBytes            = []byte("gzip")
)

type RequestURIFn func(c *routing.Context) string

func TransparentHandler(requestURIFn func(c *routing.Context) string, client *fasthttp.Client) func(c *routing.Context) error {
	return func(c *routing.Context) error {
		logrus.Debugf("proxying over - %s", c.Request.RequestURI())

		req := fasthttp.AcquireRequest()
		c.Request.Header.CopyTo(&req.Header)

		req.SetRequestURI(requestURIFn(c))

		req.SetBody(c.Request.Body())

		resp := fasthttp.AcquireResponse()
		if err := client.Do(req, resp); err != nil {
			logrus.Error(err)
			return err
		}

		resp.Header.CopyTo(&c.Response.Header)
		c.Response.SetStatusCode(resp.StatusCode())
		c.Response.SetBody(resp.Body())

		return nil
	}
}

func TransparentOverCacheHandler(requestURIFn RequestURIFn, client *fasthttp.Client, store *store.TTLCache) func(c *routing.Context) error {
	return func(c *routing.Context) (err error) {
		logrus.Debugf("proxying over - %s", c.Request.RequestURI())

		container := store.Get(string(c.Request.RequestURI()))
		if container != nil {
			c.Response.SetStatusCode(http.StatusOK)
			c.Response.SetBody(container.Raw())
			if container.IsGzipped() {
				c.Response.Header.SetBytesKV(contentEncodingHeaderBytes, gzipHeaderBytes)
			}

			return nil
		}

		req := fasthttp.AcquireRequest()
		c.Request.Header.CopyTo(&req.Header)
		req.SetRequestURI(requestURIFn(c))
		req.SetBody(c.Request.Body())

		resp := fasthttp.AcquireResponse()
		if err := client.Do(req, resp); err != nil {
			logrus.Error(err)
			return err
		}

		data := resp.Body()

		isGzipped := bytes.Equal(resp.Header.PeekBytes(contentEncodingHeaderBytes), gzipHeaderBytes)
		if isGzipped {
			c.Response.Header.SetBytesKV(contentEncodingHeaderBytes, gzipHeaderBytes)
		}

		store.Store(string(c.Request.RequestURI()), data, isGzipped)

		resp.Header.CopyTo(&c.Response.Header)
		c.Response.SetStatusCode(resp.StatusCode())
		c.Response.SetBody(data)

		return nil
	}
}
