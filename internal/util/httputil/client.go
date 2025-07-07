package httputil

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
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
	Request       any
	Response      any
	ErrorResponse any
	Cookies       []*http.Cookie
	Headers       http.Header
}
type HttpRequestOptionFn func(*HttpRequestOption) error

func WithRequest(reqBody any) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		option.Request = reqBody
		return nil
	}
}

func WithResponse(res any) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		if res == nil {
			return errors.Errorf("response must not be nil")
		}

		if reflect.ValueOf(res).Kind() != reflect.Ptr {
			return errors.Errorf("response must be ptr")
		}

		option.Response = res
		return nil
	}
}
func WithErrorReponse(e any) HttpRequestOptionFn {
	return func(option *HttpRequestOption) error {
		if e == nil {
			return errors.Errorf("error response must not be nil")
		}

		if reflect.ValueOf(e).Kind() != reflect.Ptr {
			return errors.Errorf("error response must be ptr")
		}

		option.ErrorResponse = e
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
	io.Copy(io.Discard, r)
	r.Close()
}

type Response struct {
	raw  *http.Response
	Body []byte
}

func (res *Response) Raw() (int, []byte) {
	return res.raw.StatusCode, res.Body
}
func (res *Response) JSON(onSuccess any, onError any) (int, error) {
	switch res.IsSuccess() {
	case true:
		if onSuccess != nil {
			if err := json.Unmarshal(res.Body, onSuccess); err != nil {
				logrus.Warnf("cannot unmarshal http response: %s", err)
				return 0, err
			}
		}
	case false:
		if onError != nil {
			if err := json.Unmarshal(res.Body, onError); err != nil {
				logrus.Warnf("cannot unmarshal error response: %s", err)
				return 0, err
			}
		}
	}

	return res.raw.StatusCode, nil
}
func (res *Response) IsSuccess() bool {
	return res.raw.StatusCode >= 200 && res.raw.StatusCode < 300
}
