package httpproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

/*
                   targetURL
                       |
					   v
incomming request -> proxy -> proxyed request

example

- targetURL: "example.com/api/v1"
- incoming request:
  - url: "abc.com/test/v1"
- proxyed request:
  - url: "example.com/api/v1/test/v1"

*/

// Proxy request to other server
func Proxy(targetURL *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := xid.New().String()

		// if token expired... renew token here?
		// if session expired... which action is best?
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.URL.Path = path.Join(targetURL.Path, c.Param("path"))
				req.Host = targetURL.Host

				// Add or Remove header
				// req.Header["my-header"] = []string{req.Header.Get("my-header")}
				// req.Header.Add("my-header", req.Header.Get("my-header"))
				// delete(req.Header, "My-Header")
				//

				logrus.WithField("req-id", rid).Tracef("[proxy] %s %s", req.Method, req.URL)
			},
		}
		proxy.ServeHTTP(c.Writer, c.Request)
		logrus.WithField("req-id", rid).Tracef("[proxy] reponse %d", c.Writer.Status())
	}
}
