package db

import (
	"context"
)

// DeleteFromDB : delete data from hardware table
func (mdb MockDB) DeleteFromDB(ctx context.Context, id string) error {
	return nil
}

// InsertIntoDB : insert data into hardware table
func (mdb MockDB) InsertIntoDB(ctx context.Context, data string) error {
	return nil
}

// GetByMAC : get data by machine mac
func (mdb MockDB) GetByMAC(ctx context.Context, mac string) (string, error) {
	return "", nil
}

// GetByIP : get data by machine ip
func (mdb MockDB) GetByIP(ctx context.Context, ip string) (string, error) {
	return "", nil
}

// GetByID : get data by machine id
func (mdb MockDB) GetByID(ctx context.Context, id string) (string, error) {
	return "", nil
}

// GetAll : get data for all machine
func (mdb MockDB) GetAll(fn func([]byte) error) error {
	return nil
}
