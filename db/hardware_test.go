// Package db - hardware tests
// The following tests validate database functionality, using both network and
// instance payloads. A postgres instance with the tink database schema is
// required in order to run the following tests.
package db

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// setupInsertIntoDB : inserts a single record in the hardware database.
func setupInsertIntoDB(t *testing.T) {
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if _, err := a.tinkDb.instance.Exec("delete from hardware;"); err != nil {
		t.Fail()
	}
	if err := a.tinkDb.InsertIntoDB(a.ctx, a.instanceData); err != nil {
		t.Logf("%s", a.instanceData)
		a.logger.Error(err)
		t.Fail()
	}
	if err := a.tinkDb.InsertIntoDB(a.ctx, a.hardwareData); err != nil {
		t.Logf("%s", a.hardwareData)
		a.logger.Error(err)
		t.Fail()
	}
}

// TestGetByMac : retrieves a single record based on the MAC address provided.
func TestGetByMac(t *testing.T) {
	setupInsertIntoDB(t)
	var response string
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetByMAC(a.ctx, "08:00:27:00:00:01"); err != nil {
		t.Fail()
	}
	var expected hardwarePayload
	if err = json.Unmarshal([]byte(a.hardwareData), &expected); err != nil {
		t.Fail()
	}
	var received hardwarePayload
	if err = json.Unmarshal([]byte(response), &received); err != nil {
		t.Fail()
	}
	if fmt.Sprintf("%+v", received) != fmt.Sprintf("%+v", expected) {
		t.Logf("\nRECIEVED: %+v\nEXPECTED: %+v", received, expected)
		t.Fail()
	}
}

// TestGetByIP : retrieves a single record based on the IP provided.
func TestGetByIP(t *testing.T) {
	setupInsertIntoDB(t)
	var response string
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetByIP(a.ctx, "192.168.1.5"); err != nil {
		t.Fail()
	}
	var expected hardwarePayload
	if err = json.Unmarshal([]byte(a.hardwareData), &expected); err != nil {
		t.Fail()
	}
	var received hardwarePayload
	if err = json.Unmarshal([]byte(response), &received); err != nil {
		t.Fail()
	}
	if fmt.Sprintf("%+v", received) != fmt.Sprintf("%+v", expected) {
		t.Logf("\nRECIEVED: %+v\nEXPECTED: %+v", received, expected)
		t.Fail()
	}
}

// TestGetByInstanceIP : retrieves a single hardware instances based on the IP
// provided.
func TestGetByInstanceIP(t *testing.T) {
	setupInsertIntoDB(t)
	var response string
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetByIP(a.ctx, "192.168.0.1"); err != nil {
		t.Fail()
	}
	var expected hardwarePayload
	if err = json.Unmarshal([]byte(a.instanceData), &expected); err != nil {
		t.Fail()
	}
	var received hardwarePayload
	if err = json.Unmarshal([]byte(response), &received); err != nil {
		t.Fail()
	}
	if fmt.Sprintf("%+v", received) != fmt.Sprintf("%+v", expected) {
		t.Logf("\nRECIEVED: %+v\nEXPECTED: %+v", received, expected)
		t.Logf("%s", response)
		t.Fail()
	}
}

// TestConcurrentInsertIntoDb : inserts five records concurrently and validates
// whether or not all five records are actually committed.
func TestConcurrentInsertIntoDb(t *testing.T) {
	var err error
	count := 10
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if _, err := a.tinkDb.instance.Query("delete from hardware;"); err != nil {
		t.Fail()
	}
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		var payload hardwarePayload
		mac := fmt.Sprintf("A%v:A%v:A%v:A%v:A%v:A%v", rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9))
		if err = json.Unmarshal([]byte(a.hardwareData), &payload); err != nil {
			t.Fail()
		}
		payload.Network.Interfaces[0].Dhcp.Mac = mac
		uuid, err := uuid.NewRandom()
		if err != nil {
			t.Fail()
		}
		payload.ID = uuid.String()
		b, err := json.Marshal(payload)
		t.Logf("PUSHING %s", b)
		if err != nil {
			t.Fail()
		}
		go func(data string, a *testArgs, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := a.tinkDb.InsertIntoDB(a.ctx, data); err != nil {
				a.logger.Error(err)
				t.Fail()
			}
		}(string(b), a, &wg)
	}
	wg.Wait()
	var resultRole int
	if err = a.tinkDb.instance.QueryRow("select count(*) as count from hardware;").Scan(&resultRole); err != nil {
		t.Fail()
	}
	if resultRole != count {
		t.Errorf("MACS PUSHED: %v. \n MACS RETURNED: %v", count, resultRole)
	}
}

// TestSerializedInsertIntoDb : inserts five records one at a time and validates
// if all five records are actually committed.
func TestSerializedInsertIntoDb(t *testing.T) {
	var err error
	count := 10
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if _, err := a.tinkDb.instance.Query("delete from hardware;"); err != nil {
		t.Fail()
	}
	for i := 0; i < count; i++ {
		var payload hardwarePayload
		mac := fmt.Sprintf("A%v:A%v:A%v:A%v:A%v:A%v", rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9), rand.Intn(9))
		if err = json.Unmarshal([]byte(a.hardwareData), &payload); err != nil {
			t.Fail()
		}
		payload.Network.Interfaces[0].Dhcp.Mac = mac
		uuid, err := uuid.NewRandom()
		if err != nil {
			t.Fail()
		}
		payload.ID = uuid.String()
		b, err := json.Marshal(payload)
		t.Logf("PUSHING %s", b)
		if err != nil {
			t.Fail()
		}
		if err := a.tinkDb.InsertIntoDB(a.ctx, string(b)); err != nil {
			a.logger.Error(err)
			t.Fail()
		}
	}
	var resultRole int
	if err = a.tinkDb.instance.QueryRow("select count(*) as count from hardware;").Scan(&resultRole); err != nil {
		t.Fail()
	}
	if resultRole != count {
		t.Errorf("MACS PUSHED: %v. \n MACS RETURNED: %v", count, resultRole)
	}
}

// TestGetByID : retrieves a single record based on the UUID provided.
func TestGetByID(t *testing.T) {
	setupInsertIntoDB(t)
	var response string
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetByID(a.ctx, ids.hardwareID); err != nil {
		t.Fail()
	}
	var expected hardwarePayload
	if err = json.Unmarshal([]byte(a.hardwareData), &expected); err != nil {
		t.Fail()
	}
	var received hardwarePayload
	if err = json.Unmarshal([]byte(response), &received); err != nil {
		t.Fail()
	}
	if fmt.Sprintf("%+v", received) != fmt.Sprintf("%+v", expected) {
		t.Logf("\nRECIEVED: %+v\nEXPECTED: %+v", received, expected)
		t.Fail()
	}
}

// TestDeleteFromDb : deletes a single record based on the UUID provided.
func TestDeleteFromDb(t *testing.T) {
	setupInsertIntoDB(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if err = a.tinkDb.DeleteFromDB(a.ctx, ids.hardwareID); err != nil {
		t.Fail()
	}
}
