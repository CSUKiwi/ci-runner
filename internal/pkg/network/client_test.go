package network

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	. "github.com/tluo-github/ci-runner/internal/pkg/common"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)


// mock 数据
func clientHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(r.Method, r.URL.String(),
		"Content-Type:", r.Header.Get("Content-Type"),
		"Accept:", r.Header.Get("Accept"),
		"Body:", string(body))

	switch r.URL.Path {
	case "/api/v4/test/ok":
	case "/api/v4/test/auth":
		w.WriteHeader(http.StatusForbidden)
	case "/api/v4/test/json":
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
		} else if r.Header.Get("Accept") != "application/json" {
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{\"key\":\"value\"}")
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// 测试 newclient
func TestNewClient(t *testing.T) {
	c, err := newClient(&RunnerCredentials{
		URL: "http://localhost:8080///",
	})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "http://localhost:8080/api/v4/", c.url.String())
}
// 测试错误 url
func TestInvalidUrl(t *testing.T) {
	_, err := newClient(&RunnerCredentials{
		URL: "address.com/ci///",
	})
	assert.Error(t, err)
}
// 测试网络请求
func TestClientDo(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(clientHandler))
	defer s.Close()

	c, err := newClient(&RunnerCredentials{
		URL: s.URL,
	})
	assert.NoError(t, err)
	assert.NotNil(t, c)

	statusCode, statusText, _ := c.doJSON("test/auth", "GET", http.StatusOK, nil, nil)
	assert.Equal(t, http.StatusForbidden, statusCode, statusText)

	// 定义一个 request json
	req := struct {
		Query bool `json:"query"`
	}{
		true,
	}
	// 定义一个 response json
	res := struct {
		Key string `json:"key"`
	}{}
	statusCode, statusText, _ = c.doJSON("test/json", "GET", http.StatusOK, nil, &res)
	assert.Equal(t, http.StatusBadRequest, statusCode, statusText)

	statusCode, statusText, _ = c.doJSON("test/json", "GET", http.StatusOK, &req, nil)
	assert.Equal(t, http.StatusNotAcceptable, statusCode, statusText)

	statusCode, statusText, _ = c.doJSON("test/json", "GET", http.StatusOK, &req, &res)
	assert.Equal(t, http.StatusOK, statusCode, statusText)
	assert.Equal(t, "value", res.Key, statusText)
}
// 测试编码集 mock
func charsetTestClientHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v4/with-charset":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"key\":\"value\"}")
	case "/api/v4/without-charset":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"key\":\"value\"}")
	case "/api/v4/without-json":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"key\":\"value\"}")
	case "/api/v4/invalid-header":
		w.Header().Set("Content-Type", "application/octet-stream, test, a=b")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"key\":\"value\"}")
	}
}
func TestClientHandleCharsetInContentType(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(charsetTestClientHandler))
	defer s.Close()

	c, _ := newClient(&RunnerCredentials{
		URL: s.URL,
	})

	res := struct {
		Key string `json:"key"`
	}{}

	statusCode, statusText, _ := c.doJSON("with-charset", "GET", http.StatusOK, nil, &res)
	assert.Equal(t, http.StatusOK, statusCode, statusText)

	statusCode, statusText, _ = c.doJSON("without-charset", "GET", http.StatusOK, nil, &res)
	assert.Equal(t, http.StatusOK, statusCode, statusText)

	statusCode, statusText, _ = c.doJSON("without-json", "GET", http.StatusOK, nil, &res)
	assert.Equal(t, -1, statusCode, statusText)

	statusCode, statusText, _ = c.doJSON("invalid-header", "GET", http.StatusOK, nil, &res)
	assert.Equal(t, -1, statusCode, statusText)
}

// 网络异常情况测试
type backoffTestCase struct {
	responseStatus int
	mustBackoff    bool
}

func tooManyRequestsHandler(w http.ResponseWriter, r *http.Request) {
	status, err := strconv.Atoi(r.Header.Get("responseStatus"))
	if err != nil {
		w.WriteHeader(599)
	} else {
		w.WriteHeader(status)
	}
}

func TestRequestsBackOff(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(tooManyRequestsHandler))
	defer s.Close()

	c, _ := newClient(&RunnerCredentials{
		URL: s.URL,
	})

	testCases := []backoffTestCase{
		{http.StatusCreated, false},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusOK, false},
		{http.StatusConflict, true},
		{http.StatusTooManyRequests, true},
		{http.StatusCreated, false},
		{http.StatusInternalServerError, true},
		{http.StatusTooManyRequests, true},
		{599, true},
		{499, true},
	}

	backoff := c.ensureBackoff("POST", "")
	for id, testCase := range testCases {
		t.Run(fmt.Sprintf("%d-%d", id, testCase.responseStatus), func(t *testing.T) {
			backoff.Reset()
			assert.Zero(t, backoff.Attempt())

			var body io.Reader
			headers := make(http.Header)
			headers.Add("responseStatus", strconv.Itoa(testCase.responseStatus))

			res, err := c.do("/", "POST", body, "application/json", headers)

			assert.NoError(t, err)
			assert.Equal(t, testCase.responseStatus, res.StatusCode)

			var expected float64
			if testCase.mustBackoff {
				expected = 1.0
			}
			assert.Equal(t, expected, backoff.Attempt())
		})
	}
}

