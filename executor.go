package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ansibleTask struct {
	PlaybookPath string                 `json:"playbookPath"`
	Parameters   map[string]interface{} `json:"parameters"`
}

//AnsibleTaskHandler to execute ansible task
func AnsibleTaskHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("RequestLogger").(*log.Entry)
	logger.Infof("executing ansible task")
	decoder := json.NewDecoder(r.Body)
	var m ansibleTask
	err := decoder.Decode(&m)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
	}

	ansibleParameters, err := json.Marshal(&m.Parameters)
	playbookPath := m.PlaybookPath

	err = executeAnsibleTask(playbookPath, m.Parameters, r.Context())
	if err != nil {
		httpErrorMessage(w, http.StatusInternalServerError, fmt.Sprintf("ansible execution failed with parameters. parameters: %v, playbookPath: %v", string(ansibleParameters), playbookPath), r.Context())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func httpErrorMessage(w http.ResponseWriter, statusCode int, message string, ctx context.Context) {
	logger := ctx.Value("RequestLogger").(*log.Entry)
	var errMsg errorMessage

	errMsg.Message = message
	b, err := json.Marshal(errMsg)

	if err != nil {
		logger.Errorln("Failed to encode json message")
		http.Error(w, "failed to encode json message", statusCode)
	} else {
		logger.Errorln(errMsg.Message)
		http.Error(w, string(b), statusCode)
	}
}

type errorMessage struct {
	Message string `json:"message"`
}

func executeAnsibleTask(playbookPath string, parameters map[string]interface{}, ctx context.Context) error {
	logger := ctx.Value("RequestLogger").(*log.Entry)
	playbook := &AnsiblePlaybookCmd{
		Playbook: playbookPath,
		ConnectionOptions: &AnsiblePlaybookConnectionOptions{
			Connection: "local",
		},

		Options: &AnsiblePlaybookOptions{
			ExtraVars: parameters,
		},
	}

	logger.Infof("executing ansible playbook with command %v %v %v %v", "ansible-playbook -v", playbookPath, "--extra-vars", parameters)

	err := playbook.Run(ctx)
	if err != nil {
		logger.Errorf(err.Error())
		return fmt.Errorf("ansible execution failed with parameters. paramters: %s playbookPath: %s", parameters, playbookPath)
	}

	return nil
}
