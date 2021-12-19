package proxy

import (
	"bytes"
	"net/http"

	"github.com/mikekonan/exchange-proxy/store"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	contentTypeBytes           = []byte("application/json")
	contentEncodingHeaderBytes = []byte("Content-Encoding")
	gzipHeaderBytes            = []byte("gzip")
)

type RequestURIFn func(c *routing.Context) string

func TransparentHandler(requestURIFn func(c *routing.Context) string, client *Client) func(c *routing.Context) error {
	return func(c *routing.Context) error {
		logrus.Debugf("proxying over - %s", c.Request.RequestURI())

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		c.Request.Header.CopyTo(&req.Header)

		req.SetRequestURI(requestURIFn(c))

		req.SetBody(c.Request.Body())

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)
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

func TransparentOverCacheHandler(requestURIFn RequestURIFn, client *Client, store *store.TTLCache) func(c *routing.Context) error {
	return func(c *routing.Context) (err error) {
		logrus.Debugf("proxying over - %s", c.Request.RequestURI())

		container := store.Get(string(c.Request.RequestURI()))
		if container != nil {
			c.Response.SetStatusCode(http.StatusOK)
			c.Response.SetBody(container.Raw())
			c.Response.Header.SetContentTypeBytes(contentTypeBytes)
			c.Response.Header.SetContentLength(len(container.Raw()))

			return nil
		}

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		c.Request.Header.CopyTo(&req.Header)
		req.SetRequestURI(requestURIFn(c))
		req.SetBody(c.Request.Body())

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)
		if err := client.Do(req, resp); err != nil {
			logrus.Error(err)
			return err
		}

		var data []byte

		if bytes.Equal(resp.Header.PeekBytes(contentEncodingHeaderBytes), gzipHeaderBytes) {
			data, err = resp.BodyGunzip()
			if err != nil {
				return err
			}
		} else {
			data = resp.Body()
		}

		store.Store(string(c.Request.RequestURI()), data)

		c.Response.Header.SetContentTypeBytes(contentTypeBytes)
		c.Response.Header.SetContentLength(len(data))
		c.Response.SetStatusCode(resp.StatusCode())
		c.Response.SetBody(data)

		return nil
	}
}
