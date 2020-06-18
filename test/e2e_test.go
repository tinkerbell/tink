package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
	"github.com/tinkerbell/tink/test/framework"
)

var log *logrus.Logger = framework.Log

func sourceEnvs() {
	os.Setenv("TINKERBELL_REGISTRY_USERNAME", "admin")
	os.Setenv("TINKERBELL_REGISTRY_PASSWORD", "admin123")
	os.Setenv("TINKERBELL_HOST_IP", "127.0.0.1")
	os.Setenv("TINKERBELL_NGINX_IP", "127.0.0.2")
	os.Setenv("FACILITY", "onprem")
	os.Setenv("TINKERBELL_TINK_USERNAME", "admin")
	os.Setenv("TINKERBELL_TINK_PASSWORD", "admin123")
}

func TestMain(m *testing.M) {
	log.Infoln("########Creating Setup########")
	time.Sleep(10 * time.Second)
	sourceEnvs()
	err := framework.StartStack()
	if err != nil {
		os.Exit(1)
	}
	os.Setenv("TINKERBELL_GRPC_AUTHORITY", "127.0.0.1:42113")
	os.Setenv("TINKERBELL_CERT_URL", "http://127.0.0.1:42114/cert")
	client.Setup()
	log.Infoln("########Setup Created########")

	log.Infoln("Creating hardware inventory")
	//push hardware data into hardware table
	hwData := []string{"hardware_1.json", "hardware_2.json"}
	err = framework.PushHardwareData(hwData)
	if err != nil {
		log.Errorln("Failed to push hardware inventory : ", err)
		os.Exit(2)
	}
	log.Infoln("Hardware inventory created")

	log.Infoln("########Starting Tests########")
	status := m.Run()
	log.Infoln("########Finished Tests########")
	log.Infoln("########Removing setup########")
	//err = framework.TearDown()
	if err != nil {
		os.Exit(3)
	}
	log.Infoln("########Setup removed########")
	os.Exit(status)
}

var testCases = map[string]struct {
	hardware string
	hMAC     string
	template string
	workers  int64
	expected workflow.ActionState
	ephData  string
}{
	"testWfWithWorker": {"hardware_1.json", "98:03:9b:89:d7:ba", "sample_1", 1, workflow.ActionState_ACTION_SUCCESS, `{"action_02": "data_02"}`},
	"testWfTimeout":    {"hardware_1.json", "98:03:9b:89:d7:ba", "sample_2", 1, workflow.ActionState_ACTION_TIMEOUT, `{"action_01": "data_01"}`},
	//"testWfWithMultiWorkers": {"hardware_1.json", "sample_3", 2, workflow.ActionState_ACTION_SUCCESS, `{"action_01": "data_01"}`},
}

var runTestMap = map[string]func(t *testing.T){
	"testWfWithWorker": TestWfWithWorker,
	"testWfTimeout":    TestWfTimeout,
	//"testWfWithMultiWorkers": TestWfWithMultiWorkers,
}

func TestE2E(t *testing.T) {
	for key, val := range runTestMap {
		t.Run(key, val)
	}
}
