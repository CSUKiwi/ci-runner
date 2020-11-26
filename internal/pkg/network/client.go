package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fdev-ci/ci-runner/internal/pkg/common"
	"github.com/jpillora/backoff"
	"net"
	"strings"
	"sync"

	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"
)

// 封装的 request 请求凭证包含 url,token
type requestCredentials interface {
	GetURL() string
	GetToken() string
}

var (
	dialer = net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	backOffDelayMin    = 100 * time.Millisecond
	backOffDelayMax    = 60 * time.Second
	backOffDelayFactor = 2.0
	backOffDelayJitter = true
)

type client struct {
	http.Client
	url 			*url.URL
	updateTime      time.Time
	requestBackOffs map[string]*backoff.Backoff
	lock            sync.Mutex
}

func isResponseApplicationJSON(res *http.Response) (result bool, err error) {
	contentType := res.Header.Get("Content-Type")
	mimetype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false, fmt.Errorf("Content-Type parsing error: %v", err)
	}
	if mimetype != "application/json" {
		return false, fmt.Errorf("Server should return application/json. Got: %v", contentType)
	}
	return true, nil
}

func (n *client) createTransport() {
	// create transport
	n.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Minute,
	}
	n.Timeout = common.DefaultNetworkClientTimeout
}
func (n *client) ensureConfig() {
	// create or update transport
	if n.Transport == nil {
		n.updateTime = time.Now()
		n.createTransport()
	}
}


func (n *client) ensureBackoff(method, uri string) *backoff.Backoff {
	n.lock.Lock()
	defer n.lock.Unlock()

	key := fmt.Sprintf("%s_%s", method, uri)
	if n.requestBackOffs[key] == nil {
		n.requestBackOffs[key] = &backoff.Backoff{
			Min:    backOffDelayMin,
			Max:    backOffDelayMax,
			Factor: backOffDelayFactor,
			Jitter: backOffDelayJitter,
		}
	}

	return n.requestBackOffs[key]
}

func (n *client) backoffRequired(res *http.Response) bool {
	return res.StatusCode >= 400 && res.StatusCode < 600
}

// 封装发出请求,在异常情况使用二进制指数避退算法自动重试功能
func (n *client) doBackoffRequest(req *http.Request) (res *http.Response, err error) {
	res, err = n.Do(req)
	if err != nil {
		err = fmt.Errorf("couldn't execute %v against %s: %v", req.Method, req.URL, err)
		return
	}

	backoffDelay := n.ensureBackoff(req.Method, req.RequestURI)
	if n.backoffRequired(res) {
		time.Sleep(backoffDelay.Duration())
	} else {
		backoffDelay.Reset()
	}

	return
}

// 封装网络请求
func (n *client) do(uri, method string, request io.Reader, requestType string, headers http.Header) (res *http.Response, err error) {
	url, err := n.url.Parse(uri)
	if err != nil {
		return
	}

	req, err := http.NewRequest(method, url.String(), request)
	if err != nil {
		err = fmt.Errorf("failed to create NewRequest: %v", err)
		return
	}

	if headers != nil {
		req.Header = headers
	}

	if request != nil {
		req.Header.Set("Content-Type", requestType)
	}

	res, err = n.doBackoffRequest(req)
	return
}

// 封装JSON 网络请求
func (n *client) doJSON(uri, method string, statusCode int, request interface{}, response interface{}) (int, string, *http.Response) {
	var body io.Reader

	if request != nil {
		requestBody, err := json.Marshal(request)
		if err != nil {
			return -1, fmt.Sprintf("failed to marshal project object: %v", err), nil
		}
		body = bytes.NewReader(requestBody)
	}

	headers := make(http.Header)
	if response != nil {
		headers.Set("Accept", "application/json")
	}

	res, err := n.do(uri, method, body, "application/json", headers)
	if err != nil {
		return -1, err.Error(), nil
	}
	defer res.Body.Close()
	defer io.Copy(ioutil.Discard, res.Body)

	if res.StatusCode == statusCode {
		if response != nil {
			isApplicationJSON, err := isResponseApplicationJSON(res)
			if !isApplicationJSON {
				return -1, err.Error(), nil
			}

			d := json.NewDecoder(res.Body)
			err = d.Decode(response)
			if err != nil {
				return -1, fmt.Sprintf("Error decoding json payload %v", err), nil
			}
		}
	}

	return res.StatusCode, res.Status, res
}

func fixCIURL(url string) string {
	url = strings.TrimRight(url, "/")
	return url
}

func newClient(requestCredentials requestCredentials) (c *client, err error) {
	url, err := url.Parse(fixCIURL(requestCredentials.GetURL())+ "/api/v4/")
	if err != nil {
		return
	}
	if url.Scheme != "http" {
		err = errors.New("only http scheme supported")
		return
	}
	c = &client{
		url: url,
		requestBackOffs: make(map[string]*backoff.Backoff),
	}
	return
}
