package httpserver

import (
	"context"
	"encoding/json"
	"errors"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/packethost/pkg/log"

	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"testing"

	"github.com/tinkerbell/tink/protos/hardware"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

type server struct {
	hardware.UnimplementedHardwareServiceServer
}

func (s *server) Push(ctx context.Context, in *hardware.PushRequest) (*hardware.Empty, error) {
	hw := in.Data
	if hw.Id == "" {
		err := errors.New("id must be set to a UUID, got id: " + hw.Id)
		return &hardware.Empty{}, err
	}

	_, err := json.Marshal(hw)
	if err != nil {
		logger.Error(err)
	}

	// pretend to push into the db

	return &hardware.Empty{}, err
}

func TestMain(m *testing.M) {
	os.Setenv("PACKET_ENV", "test")
	os.Setenv("PACKET_VERSION", "ignored")
	os.Setenv("ROLLBAR_TOKEN", "ignored")

	logger, _, _ = log.Init("github.com/tinkerbell/tink")

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	hardware.RegisterHardwareServiceServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Info("Server exited with error: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func TestHardwarePushHandler(t *testing.T) {

	for name, test := range handlerTests {
		t.Log(name)

		mux := grpcRuntime.NewServeMux()
		dialOpts := []grpc.DialOption{grpc.WithContextDialer(bufDialer), grpc.WithInsecure()}
		grpcEndpoint := "localhost:42113"

		err := RegisterHardwareServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, dialOpts)

		req, err := http.NewRequest("POST", "/v1/hardware", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Body = ioutil.NopCloser(strings.NewReader(test.json))
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, req)

		if status := resp.Code; status != test.status {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, test.status)
		}

		t.Log(resp.Body.String())
	}
}

var handlerTests = map[string]struct {
	id     string
	status int
	error  string
	json   string
}{
	"hardware push": {
		id:     "fde7c87c-d154-447e-9fce-7eb7bdec90c0",
		status: http.StatusOK,
		json:   hardwarePushData,
	},
	"tinkerbell no metadata": {
		id:     "363115b0-f03d-4ce5-9a15-5514193d131a",
		status: http.StatusOK,
		json:   hardwarePushDataNoMetadata,
	},
	"hardware push no id": {
		status: http.StatusBadRequest,
		json:   hardwarePushDataNoID,
	},
	"hardware push invalid json": {
		status: http.StatusBadRequest,
		json:   hardwarePushDataInvalidJSON,
	},
}

const (
	hardwarePushData = `
{
  "id": "0eba0bf8-3772-4b4a-ab9f-6ebe93b90a94",
  "metadata": {
    "test": "testing unos dos tres",
    "bonding_mode": 5,
    "custom": {
      "preinstalled_operating_system_version": {},
      "private_subnets": []
    },
    "facility": {
      "facility_code": "ewr1",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
    },
    "instance": {
      "crypted_root_password": "redacted",
      "operating_system_version": {
        "distro": "ubuntu",
        "os_slug": "ubuntu_18_04",
        "version": "18.04"
      },
      "storage": {
        "disks": [
          {
            "device": "/dev/sda",
            "partitions": [
              {
                "label": "BIOS",
                "number": 1,
                "size": 4096
              },
              {
                "label": "SWAP",
                "number": 2,
                "size": 3993600
              },
              {
                "label": "ROOT",
                "number": 3,
                "size": 0
              }
            ],
            "wipe_table": true
          }
        ],
        "filesystems": [
          {
            "mount": {
              "create": {
                "options": ["-L", "ROOT"]
              },
              "device": "/dev/sda3",
              "format": "ext4",
              "point": "/"
            }
          },
          {
            "mount": {
              "create": {
                "options": ["-L", "SWAP"]
              },
              "device": "/dev/sda2",
              "format": "swap",
              "point": "none"
            }
          }
        ]
      }
    },
    "manufacturer": {
      "id": "",
      "slug": ""
    },
    "state": ""
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "hostname": "server001",
          "ip": {
            "address": "192.168.1.5",
            "gateway": "192.168.1.1",
            "netmask": "255.255.255.248"
          },
          "lease_time": 86400,
          "mac": "00:00:00:00:00:00",
          "name_servers": [],
          "time_servers": [],
          "uefi": false
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true,
          "ipxe": {
            "contents": "#!ipxe",
            "url": "http://url/menu.ipxe"
          },
          "osie": {
            "base_url": "",
            "initrd": "",
            "kernel": "vmlinuz-x86_64"
          }
        }
      }
    ]
  }
}
`
	hardwarePushDataNoMetadata = `
{
	   "network":{
		  "interfaces":[
			 {
				"dhcp":{
				   "mac":"ec:0d:9a:c0:01:0c",
				   "hostname":"server001",
				   "lease_time":86400,
				   "arch":"x86_64",
				   "ip":{
					  "address":"192.168.1.5",
					  "netmask":"255.255.255.248",
					  "gateway":"192.168.1.1"
				   }
				},
				"netboot":{
				   "allow_pxe":true,
				   "allow_workflow":true,
				   "ipxe":{
					  "url":"http://url/menu.ipxe",
					  "contents":"#!ipxe"
				   },
				   "osie":{
					  "kernel":"vmlinuz-x86_64"
				   }
				}
			 }
		  ]
	   },
	   "id":"363115b0-f03d-4ce5-9a15-5514193d131a"
	}
`
	hardwarePushDataNoID = `
{
  "metadata": {
    "facility": {
      "facility_code": "ewr1",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
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
          "mac": "00:00:00:00:00:00",
          "uefi": false
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
`
	hardwarePushDataInvalidJSON = `
  "id": "363115b0-f03d-4ce5-9a15-5514193d131a",
  "metadata": {
    "bonding_mode": 5,
    "custom": {
      "preinstalled_operating_system_version": {},
      "private_subnets": []
    },
    "facility": {
      "facility_code": "ewr1",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
    },
    "instance": {},
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "hostname": "server001",
          "ip": {
            "address": "192.168.1.5",
            "gateway": "192.168.1.1",
            "netmask": "255.255.255.248"
          },
          "lease_time": 86400,
          "mac": "00:00:00:00:00:00",
          "name_servers": [],
          "time_servers": [],
          "uefi": false
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true,
          "ipxe": {
            "contents": "#!ipxe",
            "url": "http://url/menu.ipxe"
          },
          "osie": {
            "base_url": "",
            "initrd": "",
            "kernel": "vmlinuz-x86_64"
          }
        }
      }
    ]
  }
}
`
)
