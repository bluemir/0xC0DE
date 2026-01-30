package httputil

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"maps"
	"mime"
	"net/http"
	"reflect"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/util/retry"
)

const HttpHeaderAuthorization = "Authorization"

type IHttpClient interface {
	NewRequest(method string, url string, opts ...HttpRequestOptionFn) (*Response, error)
}

// HTTP Client with retry, req encoding, res decoding
type HttpClient struct {
	client http.Client
}

type HttpClientOptionFn func(tr *http.Transport, client *http.Client)

func NewHttpClient(opts ...HttpClientOptionFn) IHttpClient {
	tr := http.DefaultTransport.(*http.Transport).Clone()

	// default value
	tr.MaxIdleConns = 100
	tr.MaxConnsPerHost = 100
	tr.MaxIdleConnsPerHost = 100

	for _, fn := range opts {
		fn(tr, nil)
	}

	client := http.Client{
		Transport: tr,
		Timeout:   10 * time.Second, // default value
	}

	for _, fn := range opts {
		fn(nil, &client)
	}
	return &HttpClient{client}
}
func WithTimeout(d time.Duration) HttpClientOptionFn {
	return func(tr *http.Transport, client *http.Client) {
		if client == nil {
			return
		}
		client.Timeout = d
	}
}
func WithMaxIdleConns(n int) HttpClientOptionFn {
	return func(tr *http.Transport, client *http.Client) {
		if tr == nil {
			return
		}
		tr.MaxIdleConns = n
	}
}
func WithMaxIdleConnsPerHost(n int) HttpClientOptionFn {
	return func(tr *http.Transport, client *http.Client) {
		if tr == nil {
			return
		}
		tr.MaxIdleConnsPerHost = n
	}
}
func MaxConnsPerHost(n int) HttpClientOptionFn {
	return func(tr *http.Transport, client *http.Client) {
		if tr == nil {
			return
		}
		tr.MaxConnsPerHost = n
	}
}

type HttpRequestOption struct {
	Request any
	Cookies []*http.Cookie
	Headers http.Header
}
type HttpRequestOptionFn func(*HttpRequestOption) error

func WithRequest(reqBody any) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		option.Request = reqBody
		return nil
	}
}
func WithCookie(cookie []*http.Cookie) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		option.Cookies = cookie
		return nil
	}
}
func WithHeader(h http.Header) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		option.Headers = h
		return nil
	}
}

func (c *HttpClient) NewRequest(method string, url string, opts ...HttpRequestOptionFn) (*Response, error) {
	option := HttpRequestOption{
		Headers: http.Header{},
	}

	for _, fn := range opts {
		if err := fn(&option); err != nil {
			return nil, err
		}
	}

	req, err := c.makeRequest(method, url, option.Request)
	if err != nil {
		return nil, err
	}

	// copy header
	maps.Copy(req.Header, option.Headers)

	// copy cookie
	for _, cookie := range option.Cookies {
		req.AddCookie(cookie)
	}

	var res *http.Response
	if err = retry.Retry(3, func() error {
		res, err = c.client.Do(req)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	defer c.ensureReaderClosed(res.Body)

	logrus.Debugf("http request method: %s, url: %s, status code: %d", method, url, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if len(resBody) < 1 {
		return &Response{raw: res, Body: nil}, nil
	}

	return &Response{
		res,
		resBody,
	}, nil
}

func (c *HttpClient) makeRequest(method string, url string, reqObject any) (*http.Request, error) {
	if reqObject == nil {
		return http.NewRequest(method, url, nil)
	}

	reqBody, err := json.Marshal(reqObject)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *HttpClient) ensureReaderClosed(r io.ReadCloser) {
	if _, err := io.Copy(io.Discard, r); err != nil {
		logrus.Tracef("failed to discard remaining body: %v", err)
	}
	if err := r.Close(); err != nil {
		logrus.Tracef("failed to close body: %v", err)
	}
}

type Response struct {
	raw  *http.Response
	Body []byte
}

func (res *Response) Raw() (int, []byte) {
	return res.raw.StatusCode, res.Body
}
func (res *Response) ShouldBind(onSuccess any, onError any) (int, error) {
	kind, _, err := mime.ParseMediaType(res.raw.Header.Get("Content-Type"))
	if err != nil {
		// TODO consider as plain text?
		return 0, errors.WithStack(err)
	}

	switch kind {
	case "application/json", "text/json":
		return res.ShouldBindWith(onSuccess, onError, json.Unmarshal)
	case "text/xml":
		return res.ShouldBindWith(onSuccess, onError, xml.Unmarshal)
	case "text/plain":
		return res.ShouldBindWith(onSuccess, onError, TextBinder)
	default:
		return res.ShouldBindWith(onSuccess, onError, TextBinder)
	}
}
func (res *Response) ShouldBindWith(onSuccess, onError any, binder func(buf []byte, v any) error) (int, error) {
	switch res.IsSuccess() {
	case true:
		if onSuccess != nil {
			if err := binder(res.Body, onSuccess); err != nil {
				logrus.Warnf("cannot unmarshal http response: %s", err)
				return 0, err
			}
		}
	case false:
		if onError != nil {
			if err := binder(res.Body, onError); err != nil {
				logrus.Warnf("cannot unmarshal error response: %s", err)
				return 0, err
			}
		}
	}
	return res.raw.StatusCode, nil
}
func TextBinder(buf []byte, v any) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.Errorf("Must be Ptr")
	}

	switch value := v.(type) {
	case *string:
		*value = string(buf)
	case *[]byte:
		*value = buf
	default:
		return errors.Errorf("not implements")
	}
	return nil
}
func (res *Response) IsSuccess() bool {
	return res.raw.StatusCode >= 200 && res.raw.StatusCode < 300
}
