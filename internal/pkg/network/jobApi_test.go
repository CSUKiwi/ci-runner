package network

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	. "github.com/fdev-ci/ci-runner/internal/pkg/common"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClients(t *testing.T) {
	c := NewCiApiClient()
	c1, _ := c.getClient(&RunnerCredentials{
		URL: "http://test/",
	})
	c2, _ := c.getClient(&RunnerCredentials{
		URL: "http://test2/",
	})

	assert.NotEqual(t, c1, c2)
}

// mock requestjob 处理



func getRequestJobResponse() (res map[string]interface{}) {
	jobToken := "job-token"
	res = make(map[string]interface{})
	res["id"] = 10
	res["token"] = jobToken

	return
}
func testRequestJobHandler(w http.ResponseWriter, r *http.Request, t *testing.T) {
	if r.URL.Path != "/api/v4/jobs/request" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	var req map[string]interface{}
	err = json.Unmarshal(body, &req)
	assert.NoError(t, err)

	switch req["token"].(string) {
	case "valid":
	case "no-jobs":
		w.Header().Add("X-GitLab-Last-Update", "a nice timestamp")
		w.WriteHeader(http.StatusNoContent)
		return
	case "invalid":
		w.WriteHeader(http.StatusForbidden)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := json.Marshal(getRequestJobResponse())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(output)
	t.Logf("JobRequest response: %s\n", output)
}

func TestRequestJob(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testRequestJobHandler(w, r, t)
	}))
	defer s.Close()

	validToken := RunnerConfig{
		RunnerCredentials: RunnerCredentials{
			URL:   s.URL,
			Token: "valid",
		},
	}

	noJobsToken := RunnerConfig{
		RunnerCredentials: RunnerCredentials{
			URL:   s.URL,
			Token: "no-jobs",
		},
	}

	invalidToken := RunnerConfig{
		RunnerCredentials: RunnerCredentials{
			URL:   s.URL,
			Token: "invalid",
		},
	}

	c := NewCiApiClient()

	res, ok := c.RequestJob(validToken)
	if assert.NotNil(t, res) {
		assert.NotEmpty(t, res.ID)
	}
	assert.True(t, ok)
	assert.Equal(t, 10, res.ID)


	res, ok = c.RequestJob(noJobsToken)
	assert.Nil(t, res)
	assert.True(t, ok, "If no jobs, runner is healthy")


	res, ok = c.RequestJob(invalidToken)
	assert.Nil(t, res)
	assert.False(t, ok, "If token is invalid, the runner is unhealthy")

}
