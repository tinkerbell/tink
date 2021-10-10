package mock

import (
	"context"
)

// DeleteFromDB : delete data from hardware table.
func (d DB) DeleteFromDB(_ context.Context, _ string) error {
	return nil
}

// InsertIntoDB : insert data into hardware table.
func (d DB) InsertIntoDB(_ context.Context, _ string) error {
	return nil
}

// GetByMAC : get data by machine mac.
func (d DB) GetByMAC(_ context.Context, _ string) (string, error) {
	return "", nil
}

// GetByIP : get data by machine ip.
func (d DB) GetByIP(_ context.Context, _ string) (string, error) {
	return "", nil
}

// GetByID : get data by machine id.
func (d DB) GetByID(_ context.Context, _ string) (string, error) {
	return "", nil
}

// GetAll : get data for all machine.
func (d DB) GetAll(_ func([]byte) error) error {
	return nil
}
