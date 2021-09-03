package mock

import (
	"context"
)

// DeleteFromDB : delete data from hardware table.
func (d DB) DeleteFromDB(ctx context.Context, id string) error {
	return nil
}

// InsertIntoDB : insert data into hardware table.
func (d DB) InsertIntoDB(ctx context.Context, data string) error {
	return nil
}

// GetByMAC : get data by machine mac.
func (d DB) GetByMAC(ctx context.Context, mac string) (string, error) {
	return "", nil
}

// GetByIP : get data by machine ip.
func (d DB) GetByIP(ctx context.Context, ip string) (string, error) {
	return "", nil
}

// GetByID : get data by machine id.
func (d DB) GetByID(ctx context.Context, id string) (string, error) {
	return "", nil
}

// GetAll : get data for all machine.
func (d DB) GetAll(fn func([]byte) error) error {
	return nil
}
