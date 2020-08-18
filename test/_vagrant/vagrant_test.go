package vagrant_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	vagrant "github.com/tinkerbell/tink/test/_vagrant"
	"github.com/tinkerbell/tink/util"
)

func TestVagrantSetupGuide(t *testing.T) {
	ctx := context.Background()

	machine, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("provisioner"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := machine.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	_, err = machine.Exec(ctx, "cd /vagrant/deploy && source ../envrc && docker-compose up -d")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker pull hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker tag hello-world 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker push 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	for ii := 0; ii < 5; ii++ {
		resp, err := http.Get("http://localhost:42114/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			if err != nil {
				t.Logf("err tinkerbell healthcheck... retrying: %s", err)
			} else {
				t.Logf("err tinkerbell healthcheck... expected status code 200 got %d retrying", resp.StatusCode)
			}
			time.Sleep(10 * time.Second)
		}
	}

	t.Log("Tinkerbell is up and running")

	os.Setenv("TINKERBELL_CERT_URL", "http://127.0.0.1:42114/cert")
	os.Setenv("TINKERBELL_GRPC_AUTHORITY", "127.0.0.1:42113")
	client.Setup()
	_, err = client.HardwareClient.All(ctx, &hardware.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	hwDataFile := "data.json"

	err = registerHardwares(ctx, hwDataFile)
	if err != nil {
		t.Fatal(err)
	}

	templateID, err := registerTemplates(ctx, "hello-world.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("templateID: %s", templateID)

	workflowID, err := createWorkflow(ctx, templateID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("WorkflowID: %s", workflowID)

	os.Setenv("VAGRANT_WORKER_GUI", "false")
	worker, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("worker"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
		vagrant.RunAsync(),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := worker.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	for iii := 0; iii < 30; iii++ {
		events, err := client.WorkflowClient.ShowWorkflowEvents(ctx, &workflow.GetRequest{
			Id: workflowID,
		})
		if err != nil {
			t.Fatal(err)
		}
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			if event.ActionName == "hello_world" && event.ActionStatus == workflow.ActionState_ACTION_SUCCESS {
				t.Logf("event %s SUCCEEDED as expected", event.ActionName)
				return
			}
		}
		time.Sleep(10 * time.Second)
	}
	t.Fatal("Workflow never got to a complite state or it didn't make it on time (10m)")
}
func TestOneTimeoutWorkflow(t *testing.T) {
	ctx := context.Background()

	machine, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("provisioner"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := machine.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	_, err = machine.Exec(ctx, "cd /vagrant/deploy && source ../envrc && docker-compose up -d")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker pull hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker pull bash")
	if err != nil {
		t.Fatal(err)
	}
	_, err = machine.Exec(ctx, "docker tag hello-world 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker tag bash 192.168.1.1/bash")
	if err != nil {
		t.Fatal(err)
	}
	_, err = machine.Exec(ctx, "docker push 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker push 192.168.1.1/bash")
	if err != nil {
		t.Fatal(err)
	}

	for ii := 0; ii < 5; ii++ {
		resp, err := http.Get("http://localhost:42114/_packet/healthcheck")
		if err != nil || resp.StatusCode != http.StatusOK {
			if err != nil {
				t.Logf("err tinkerbell healthcheck... retrying: %s", err)
			} else {
				t.Logf("err tinkerbell healthcheck... expected status code 200 got %d retrying", resp.StatusCode)
			}
			time.Sleep(10 * time.Second)
		}
	}

	t.Log("Tinkerbell is up and running")

	os.Setenv("TINKERBELL_CERT_URL", "http://127.0.0.1:42114/cert")
	os.Setenv("TINKERBELL_GRPC_AUTHORITY", "127.0.0.1:42113")
	client.Setup()
	_, err = client.HardwareClient.All(ctx, &hardware.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	hwDataFile := "data.json"

	err = registerHardwares(ctx, hwDataFile)
	if err != nil {
		t.Fatal(err)
	}

	templateID, err := registerTemplates(ctx, "timeout.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("templateID: %s", templateID)

	workflowID, err := createWorkflow(ctx, templateID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("WorkflowID: %s", workflowID)

	os.Setenv("VAGRANT_WORKER_GUI", "false")
	worker, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("worker"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
		vagrant.RunAsync(),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := worker.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	for iii := 0; iii < 30; iii++ {
		events, err := client.WorkflowClient.ShowWorkflowEvents(ctx, &workflow.GetRequest{
			Id: workflowID,
		})
		if err != nil {
			t.Fatal(err)
		}
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			if event.ActionName == "hello_world" && event.ActionStatus == workflow.ActionState_ACTION_SUCCESS {
				t.Logf("action %s SUCCESSFULL as expected", event.ActionName)
				continue
			}
			if event.ActionName == "sleep-till-timeout" && event.ActionStatus == workflow.ActionState_ACTION_TIMEOUT {
				t.Logf("action %s TIMEDOUT as expected", event.ActionName)
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatal("Workflow never got to a complite state or it didn't make it on time (5m)")
}

func TestOneFailedWorkflow(t *testing.T) {
	ctx := context.Background()

	machine, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("provisioner"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := machine.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	_, err = machine.Exec(ctx, "cd /vagrant/deploy && source ../envrc && docker-compose up -d")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker pull hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker pull bash")
	if err != nil {
		t.Fatal(err)
	}
	_, err = machine.Exec(ctx, "docker tag hello-world 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker tag bash 192.168.1.1/bash")
	if err != nil {
		t.Fatal(err)
	}
	_, err = machine.Exec(ctx, "docker push 192.168.1.1/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	_, err = machine.Exec(ctx, "docker push 192.168.1.1/bash")
	if err != nil {
		t.Fatal(err)
	}

	for ii := 0; ii < 5; ii++ {
		resp, err := http.Get("http://localhost:42114/_packet/healthcheck")
		if err != nil || resp.StatusCode != http.StatusOK {
			if err != nil {
				t.Logf("err tinkerbell healthcheck... retrying: %s", err)
			} else {
				t.Logf("err tinkerbell healthcheck... expected status code 200 got %d retrying", resp.StatusCode)
			}
			time.Sleep(10 * time.Second)
		}
	}

	t.Log("Tinkerbell is up and running")

	os.Setenv("TINKERBELL_CERT_URL", "http://127.0.0.1:42114/cert")
	os.Setenv("TINKERBELL_GRPC_AUTHORITY", "127.0.0.1:42113")
	client.Setup()
	_, err = client.HardwareClient.All(ctx, &hardware.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	hwDataFile := "data.json"

	err = registerHardwares(ctx, hwDataFile)
	if err != nil {
		t.Fatal(err)
	}

	templateID, err := registerTemplates(ctx, "failedWorkflow.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("templateID: %s", templateID)

	workflowID, err := createWorkflow(ctx, templateID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("WorkflowID: %s", workflowID)

	os.Setenv("VAGRANT_WORKER_GUI", "false")
	worker, err := vagrant.Up(ctx,
		vagrant.WithLogger(t.Logf),
		vagrant.WithMachineName("worker"),
		vagrant.WithWorkdir("../../deploy/vagrant"),
		vagrant.RunAsync(),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := worker.Destroy(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	for iii := 0; iii < 30; iii++ {
		events, err := client.WorkflowClient.ShowWorkflowEvents(ctx, &workflow.GetRequest{
			Id: workflowID,
		})
		if err != nil {
			t.Fatal(err)
		}
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			if event.ActionName == "hello_world" && event.ActionStatus == workflow.ActionState_ACTION_SUCCESS {
				t.Logf("action %s SUCCESSFULL as expected", event.ActionName)
				continue
			}
			if event.ActionName == "sleep-till-timeout" && event.ActionStatus == workflow.ActionState_ACTION_FAILED {
				t.Logf("action %s FAILED as expected", event.ActionName)
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatal("Workflow never got to a complite state or it didn't make it on time (5m)")
}

func createWorkflow(ctx context.Context, templateID string) (string, error) {
	res, err := client.WorkflowClient.CreateWorkflow(ctx, &workflow.CreateRequest{
		Template: templateID,
		Hardware: `{"device_1":"08:00:27:00:00:01"}`,
	})
	if err != nil {
		return "", err
	}
	return res.Id, nil
}

func readData(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return []byte(""), err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}

// push hardware data through file
func registerHardwares(ctx context.Context, hwDatafile string) error {
	//for _, hwFile := range hwDataFiles {
	//filepath := "../data/hardware/" + hwFile
	data, err := readData(hwDatafile)
	if err != nil {
		return err
	}
	hw := util.HardwareWrapper{Hardware: &hardware.Hardware{}}
	if err := json.Unmarshal(data, &hw); err != nil {
		return err
	}
	_, err = client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: hw.Hardware})
	if err != nil {
		return err
	}
	//}
	return nil
}

func registerTemplates(ctx context.Context, templateFile string) (string, error) {
	data, err := readData(templateFile)
	if err != nil {
		return "", err
	}
	name := strings.SplitN(templateFile, ".", -1)[0]
	resp, err := client.TemplateClient.CreateTemplate(ctx, &template.WorkflowTemplate{Name: name, Data: string(data)})
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}
