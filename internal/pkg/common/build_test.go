package common

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestBuildRun(t *testing.T) {
	coronaVirusJSON := `{
        "name" : "covid-11",
        "metadata" : {
			"Project_id": 11
		}
    }`
	var build Build
	err := json.Unmarshal([]byte(coronaVirusJSON),&build)
	if err != nil {
		logrus.Warn(fmt.Sprintf("json unmarshall faild: %v", err))
		// 出队列
	}

	// Reading each value by its key
	fmt.Println(build.MetaData["Project_id"])
}
