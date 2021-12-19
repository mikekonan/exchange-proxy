package proxy

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	strLocation = []byte("Location")
)

type Client struct {
	fasthttp.Client
}

func (c *Client) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	for {
		if err := c.Client.Do(req, resp); err != nil {
			return err
		}

		statusCode := resp.Header.StatusCode()
		if statusCode != fasthttp.StatusMovedPermanently &&
			statusCode != fasthttp.StatusFound &&
			statusCode != fasthttp.StatusSeeOther &&
			statusCode != fasthttp.StatusTemporaryRedirect &&
			statusCode != fasthttp.StatusPermanentRedirect {
			break
		}

		location := resp.Header.PeekBytes(strLocation)
		if len(location) == 0 {
			return fmt.Errorf("redirect with missing Location header")
		}

		u := req.URI()
		u.UpdateBytes(location)

		resp.Header.VisitAllCookie(func(key, value []byte) {
			c := fasthttp.AcquireCookie()
			defer fasthttp.ReleaseCookie(c)

			if err := c.ParseBytes(value); err != nil {
				logrus.Fatal(err)
			}

			if expire := c.Expire(); expire != fasthttp.CookieExpireUnlimited && expire.Before(time.Now()) {
				req.Header.DelCookieBytes(key)
			} else {
				req.Header.SetCookieBytesKV(key, c.Value())
			}
		})
	}

	return nil
}
