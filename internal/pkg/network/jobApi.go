package network

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tluo-github/ci-runner/internal/pkg/common"
	"net/http"
	"runtime"
	"sync"
)
const clientError = -100

type CiApiClient struct {
	clients map[string]*client
	lock    sync.Mutex
}

func (n *CiApiClient) getClient(credentials requestCredentials)(c *client, err error)   {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.clients == nil{
		n.clients = make(map[string] * client)
	}
	key := fmt.Sprintf("%s_%s", credentials.GetURL(), credentials.GetToken())
	c = n.clients[key]
	if c == nil {
		c, err = newClient(credentials)
		if err != nil {
			return
		}
		n.clients[key] = c
	}

	return
}

func (n *CiApiClient) doJSON(credentials requestCredentials, method, uri string, statusCode int, request interface{}, response interface{}) (int, string, *http.Response) {
	c, err := n.getClient(credentials)
	if err != nil {
		return clientError, err.Error(), nil
	}

	return c.doJSON(uri, method, statusCode, request, response)
}
func (n *CiApiClient) getRunnerVersion(config common.RunnerConfig) common.VersionInfo {
	info := common.VersionInfo{
		Name:         common.NAME,
		Version:      common.VERSION,
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
		Executor:     config.Executor,
	}
	return info
}

func (n *CiApiClient) RequestJob(config common.RunnerConfig)(*common.JobResponse, bool)  {

	var response common.JobResponse
	request := common.JobRequest{
		Info:  n.getRunnerVersion(config),
		Token: config.Token,
	}

	result, statusText, _ := n.doJSON(&config.RunnerCredentials, "POST", "jobs/request", http.StatusCreated, &request, &response)

	switch result {
	case http.StatusCreated:
		logrus.WithFields(logrus.Fields{
			"job":      response.ID,
		}).Println("Checking for jobs...", "received")
		return &response, true
	case http.StatusForbidden:
		logrus.Errorln("Checking for jobs...", "forbidden")
		return nil, false
	case http.StatusNoContent:
		logrus.Debugln("Checking for jobs...", "nothing")
		return nil, true
	case clientError:
		logrus.WithField("status", statusText).Errorln("Checking for jobs...", "error")
		return nil, false
	default:
		logrus.WithField("status", statusText).Warningln("Checking for jobs...", "failed")
		return nil, true
	}
}
func NewCiApiClient() *CiApiClient {
	return &CiApiClient{}
}

