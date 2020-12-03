// Package db - common test
// The following file contains all the common functions for testing.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/packethost/pkg/log"
)

var (
	ids = testID{
		instanceID: "da013837-a30c-237a-adfc-2d9ad69d5000",
		hardwareID: "ce2e62ed-826f-4485-a39f-a82bb74338e2",
		templateID: "da013837-a30c-237a-adfc-2d9ad69d5011",
		workflowID: "b356eede-22c1-11eb-8c17-0242ac120004",
	}
)

type hardwarePayload struct {
	ID       string `json:"id"`
	Metadata struct {
		Facility struct {
			FacilityCode string `json:"facility_code"`
		} `json:"facility"`
		Instance struct {
			IPAddresses []struct {
				Address string `json:"address"`
			} `json:"ip_addresses"`
		} `json:"instance"`
		State string `json:"state"`
	} `json:"metadata"`
	Network struct {
		Interfaces []struct {
			Dhcp struct {
				Arch string `json:"arch"`
				IP   struct {
					Address string `json:"address"`
					Gateway string `json:"gateway"`
					Netmask string `json:"netmask"`
				} `json:"ip"`
				Mac  string `json:"mac"`
				Uefi bool   `json:"uefi"`
			} `json:"dhcp"`
			Netboot struct {
				AllowPxe      bool `json:"allow_pxe"`
				AllowWorkflow bool `json:"allow_workflow"`
			} `json:"netboot"`
		} `json:"interfaces"`
	} `json:"network"`
}

type testArgs struct {
	host         string
	port         int32
	tinkDb       *TinkDB
	hardwareID   string
	user         string
	password     string
	dbname       string
	psqlInfo     string
	logger       log.Logger
	ctx          context.Context
	instanceData string
	hardwareData string
	// templateData - referenced in workflow.Workflow. Formatted as yaml.
	templateData        string
	templateDataUpdated string
	workflowData        Workflow
}

type testID struct {
	workflowID string
	instanceID string
	hardwareID string
	templateID string
}

func (a *testArgs) setLogger(t *testing.T) {
	var err error
	if a.logger, err = log.Init("github.com/tinkerbell/tink"); err != nil {
		t.Error(err)
	}
}

func (a *testArgs) setPsqlInfo() {
	a.psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		a.host, a.port, a.user, a.password, a.dbname)
	a.logger.Info(a.psqlInfo)
}

func newTestParams(t *testing.T, ids testID) *testArgs {
	var err error

	hardwareData := `{
    "id": "` + ids.hardwareID + `",
    "metadata": {
      "facility": {
        "facility_code": "onprem"
      },
      "instance": {
        "ip_addresses": [{
          "address": ""
        }]
      },
      "state": ""
    },
    "network": {
      "interfaces": [{
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
      }]
    }
  }`
	a := &testArgs{
		ctx:        context.Background(),
		host:       "localhost",
		port:       5432,
		tinkDb:     nil,
		user:       "tinkerbell",
		password:   "tinkerbell",
		dbname:     "tinkerbell",
		hardwareID: ids.hardwareID,
		workflowData: Workflow{
			State:    0,
			ID:       ids.workflowID,
			Hardware: hardwareData,
			Template: ids.templateID,
		},
		hardwareData: hardwareData,
		instanceData: ` {
    "id": "` + ids.instanceID + `",
    "instance": {
      "ip_addresses": [
      {
        "address": "192.168.0.1"
      }
      ]
    }
    }`,
		templateData: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
- name: "hello world"
  worker: "192.168.0.1"
  actions:
  - name: "hello_world"
    image: hello-world
    timeout: 60`,
		templateDataUpdated: `
version: "0.2"
name: hello_world_workflow
global_timeout: 600
tasks:
- name: "hello world"
  worker: "192.168.0.1"
  actions:
  - name: "hello_world"
    image: hello-world
    timeout: 60`,
	}
	////////////////////////////////////////////////////////////////////////////
	a.setLogger(t)
	a.setPsqlInfo()
	a.tinkDb = Connect(a.logger)
	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		panic(err)
	}
	return a
}
