package vagrant_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/pkg"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	vagrant "github.com/tinkerbell/tink/test/_vagrant"
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
	err = registerHardware(ctx)
	if err != nil {
		t.Fatal(err)
	}

	templateID, err := registerTemplate(ctx)
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
			if event.ActionName == "hello_world" && event.ActionStatus == workflow.ActionState_ACTION_STATE_SUCCESS {
				t.Logf("event %s SUCCEEDED as expected", event.ActionName)
				return
			}
		}
		time.Sleep(10 * time.Second)
	}
	t.Fatal("Workflow never got to a complite state or it didn't make it on time (10m)")
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

func registerTemplate(ctx context.Context) (string, error) {
	resp, err := client.TemplateClient.CreateTemplate(ctx, &template.WorkflowTemplate{
		Name: "hello-world",
		Data: `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
      - name: "hello_world"
        image: hello-world
        timeout: 60`,
	})
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func registerHardware(ctx context.Context) error {
	data := []byte(`{
  "id": "ce2e62ed-826f-4485-a39f-a82bb74338e2",
  "metadata": {
    "facility": {
      "facility_code": "onprem"
    },
    "instance": {},
    "state": ""
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "ip": {
            "address": "192.168.1.5",
            "gateway": "192.168.1.1",
            "netmask": "255.255.255.248"
          },
          "mac": "08:00:27:00:00:01",
          "uefi": false
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}`)
	hw := pkg.HardwareWrapper{Hardware: &hardware.Hardware{}}
	err := json.Unmarshal(data, &hw)
	if err != nil {
		return err
	}
	_, err = client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: hw.Hardware})
	return err
}
