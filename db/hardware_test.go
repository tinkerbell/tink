package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/pkg"
	"github.com/tinkerbell/tink/protos/hardware"
)

func TestCreateHardware(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		// Name identifies the single test in a table test scenario
		Name string
		// InputAsync if set to true inserts all the input concurrently
		InputAsync bool
		// Input is a hardware that will be used to pre-populate the database
		Input []*hardware.Hardware
		// Expectation is the function used to apply the assertions.
		// You can use it to validate if the Input are created as you expect
		Expectation func(*testing.T, []*hardware.Hardware, *db.TinkDB)
		// ExpectedErr is used to check for error during
		// CreateTemplate execution. If you expect a particular error
		// and you want to assert it, you can use this function
		ExpectedErr func(*testing.T, error)
	}{
		{
			Name:  "create-single-hardware",
			Input: []*hardware.Hardware{readHardwareData("./testdata/hardware.json")},
			Expectation: func(t *testing.T, input []*hardware.Hardware, tinkDB *db.TinkDB) {
				data, err := tinkDB.GetByID(ctx, input[0].Id)
				if err != nil {
					t.Error(err)
				}
				hw := &hardware.Hardware{}
				if err := json.Unmarshal([]byte(data), hw); err != nil {
					t.Error(err)
				}
				if dif := cmp.Diff(input[0], hw, cmp.Comparer(hardwareComparer)); dif != "" {
					t.Errorf(dif)
				}
			},
		},
		{
			Name: "two-hardware-with-same-mac",
			Input: []*hardware.Hardware{
				func() *hardware.Hardware {
					hw := readHardwareData("./testdata/hardware.json")
					hw.Id = uuid.New().String()
					return hw
				}(),
				func() *hardware.Hardware {
					hw := readHardwareData("./testdata/hardware.json")
					hw.Id = uuid.New().String()
					return hw
				}(),
			},
			ExpectedErr: func(t *testing.T, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "conflicting hardware MAC address 08:00:27:00:00:01 provided with hardware data/info") {
					t.Errorf("\nexpected err: %s\ngot: %s", "conflicting hardware MAC address 08:00:27:00:00:01 provided with hardware data/info", err)
				}
			},
		},
		{
			Name: "update-on-create",
			Input: []*hardware.Hardware{
				func() *hardware.Hardware {
					hw := readHardwareData("./testdata/hardware.json")
					hw.Id = "d71b659c-3db8-404e-be0e-2fb3c2a482bd"
					return hw
				}(),
				func() *hardware.Hardware {
					hw := readHardwareData("./testdata/hardware.json")
					hw.Id = "d71b659c-3db8-404e-be0e-2fb3c2a482bd"
					hw.Network.Interfaces[0].Dhcp.Hostname = "updated-hostname"
					return hw
				}(),
			},
			Expectation: func(t *testing.T, input []*hardware.Hardware, tinkDB *db.TinkDB) {
				data, err := tinkDB.GetByID(ctx, input[0].Id)
				if err != nil {
					t.Error(err)
				}
				hw := &hardware.Hardware{}
				if err := json.Unmarshal([]byte(data), hw); err != nil {
					t.Error(err)
				}
				hostName := hw.Network.Interfaces[0].Dhcp.Hostname
				if hostName != "updated-hostname" {
					t.Errorf("expected hostname to be \"%s\", got \"%s\"", "updated-hostname", hostName)
				}
			},
		},
		{
			Name:       "create-stress-test",
			InputAsync: true,
			Input: func() []*hardware.Hardware {
				input := []*hardware.Hardware{}
				for ii := 0; ii < 10; ii++ {
					hw := readHardwareData("./testdata/hardware.json")
					hw.Id = uuid.New().String()
					hw.Network.Interfaces[0].Dhcp.Mac = strings.Replace(hw.Network.Interfaces[0].Dhcp.Mac, "00", fmt.Sprintf("0%d", ii), 1)
				}
				return input
			}(),
			Expectation: func(t *testing.T, input []*hardware.Hardware, tinkDB *db.TinkDB) {
				count := 0
				err := tinkDB.GetAll(func(b []byte) error {
					count = count + 1
					return nil
				})
				if err != nil {
					t.Error(err)
				}
				if len(input) != count {
					t.Errorf("expected %d hardwares stored in the database but we got %d", len(input), count)
				}
			},
			ExpectedErr: func(t *testing.T, err error) {
				if err != nil {
					t.Error(err)
				}
			},
		},
	}

	for _, s := range tests {
		t.Run(s.Name, func(t *testing.T) {
			t.Parallel()
			_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
				ApplyMigration: true,
			})
			defer func() {
				err := cl()
				if err != nil {
					t.Error(err)
				}
			}()

			var wg sync.WaitGroup
			wg.Add(len(s.Input))
			for _, hw := range s.Input {
				if s.InputAsync {
					go func(ctx context.Context, tinkDB *db.TinkDB, hw *hardware.Hardware) {
						defer wg.Done()
						err := createHardware(ctx, tinkDB, hw)
						if err != nil {
							s.ExpectedErr(t, err)
						}
					}(ctx, tinkDB, hw)
				} else {
					wg.Done()
					err := createHardware(ctx, tinkDB, hw)
					if err != nil {
						s.ExpectedErr(t, err)
					}
				}
			}
			wg.Wait()
			s.Expectation(t, s.Input, tinkDB)
		})
	}
}

func TestDeleteHardware(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, tinkDB, cl := NewPostgresDatabaseClient(t, ctx, NewPostgresDatabaseRequest{
		ApplyMigration: true,
	})
	defer func() {
		err := cl()
		if err != nil {
			t.Error(err)
		}
	}()

	hw := readHardwareData("./testdata/hardware.json")
	err := createHardware(ctx, tinkDB, hw)
	if err != nil {
		t.Error(err)
	}
	err = tinkDB.DeleteFromDB(ctx, hw.Id)
	if err != nil {
		t.Error(err)
	}

	count := 0
	err = tinkDB.GetAll(func(b []byte) error {
		count = count + 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if count != 0 {
		t.Errorf("expected 0 hardwares stored in the database after delete, but we got %d", count)
	}
}

func readHardwareData(file string) *hardware.Hardware {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	var hw pkg.HardwareWrapper
	err = json.Unmarshal([]byte(data), &hw)
	if err != nil {
		panic(err)
	}
	return hw.Hardware
}

func createHardware(ctx context.Context, db *db.TinkDB, hw *hardware.Hardware) error {
	data, err := json.Marshal(hw)
	if err != nil {
		return err
	}
	return db.InsertIntoDB(ctx, string(data))
}
