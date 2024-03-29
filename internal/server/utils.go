package server

import (
	"crypto"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
)

func html(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		markHTML(c)
		c.HTML(http.StatusOK, path, c)
	}
}
func redirect(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, path)
	}
}

func getEtag() string {
	hashed := crypto.SHA512.New()

	io.WriteString(hashed, buildinfo.AppName)
	io.WriteString(hashed, buildinfo.Version)
	io.WriteString(hashed, buildinfo.BuildTime)

	return hex.EncodeToString(hashed.Sum(nil))[:20]
}
func staticCache() func(c *gin.Context) {
	etag := getEtag()
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, max-age=86400")
		c.Header("ETag", etag)

		if match := c.GetHeader("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				c.Status(http.StatusNotModified)
				c.Abort()
				return
			}
		}

		c.Request.Header.Del("If-Modified-Since") // only accept etag
	}
}
func (server *Server) AbortIfHasPrefix(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, prefix) {
			c.Abort()
		}
	}
}

func (server *Server) proxy(targetURL *url.URL) gin.HandlerFunc {
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

func fixURL(c *gin.Context) {
	url := location.Get(c)

	// QUESTION is it right?
	c.Request.URL.Scheme = url.Scheme
	c.Request.URL.Host = url.Host
}
func markAPI(c *gin.Context) {
	c.SetAccepted("application/json")
}
func markHTML(c *gin.Context) {
	c.SetAccepted("text/html")
}
