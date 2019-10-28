package grpcserver

import (
	"context"
	"encoding/json"

	"github.com/packethost/rover/db"
	"github.com/packethost/rover/protos/target"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TARGETS GRPC ENDPOINTS START
func (s *server) CreateTargets(ctx context.Context, in *target.PushRequest) (*target.UUID, error) {
	logger.Info("create Targets")

	type machine map[string]string

	var h struct {
		ID      string
		Targets map[string]machine
	}

	uuid := uuid.NewV4().String()
	err := json.Unmarshal([]byte(in.Data), &h)
	if err != nil {
		err = errors.Wrap(err, "unmarshal json")
		logger.Error(err)
		return &target.UUID{}, err
	}
	h.ID = uuid
	if h.ID == "" {
		err = errors.New("id must be set to a UUID")
		logger.Error(err)
		return &target.UUID{}, err
	}

	var fn func() error
	msg := "inserting into targets DB"
	fn = func() error { return db.InsertIntoTargetDB(ctx, s.db, in.Data, uuid) }
	logger.Info(msg)

	err = fn()

	logger.Info("done " + msg)
	logger.With("id", h.ID).Info("data pushed")
	if err != nil {
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &target.UUID{Uuid: uuid}, err
}

// This will give you the targets which belongs to the input id
func (s *server) TargetByID(ctx context.Context, in *target.GetRequest) (*target.Targets, error) {
	j, err := db.TargetsByID(ctx, s.db, in.ID)
	if err != nil {
		return &target.Targets{}, err
	}
	return &target.Targets{JSON: j}, nil
}

// This will update the targets values which belongs to the input id and data
func (s *server) UpdateTargetByID(ctx context.Context, in *target.UpdateRequest) (*target.Empty, error) {
	logger.Info("Update Targets")

	type machine map[string]string
	var t struct {
		Targets map[string]machine
	}

	err := json.Unmarshal([]byte(in.Data), &t)
	if err != nil {
		err = errors.Wrap(err, "unmarshal json")
		logger.Error(err)
		return &target.Empty{}, err
	}
	if in.ID == "" {
		err = errors.New("id must be set to an existing UUID")
		logger.Error(err)
		return &target.Empty{}, err
	}

	var fn func() error
	msg := "Updating into targets DB for ID : "
	fn = func() error { return db.InsertIntoTargetDB(ctx, s.db, in.Data, in.ID) }
	logger.Info(msg + in.ID)

	err = fn()

	logger.Info("done " + msg + in.ID)
	logger.With("id", in.ID).Info("data pushed")
	if err != nil {
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &target.Empty{}, err
}

// This will delete (soft delete) the targets which belongs to the input id
func (s *server) DeleteTargetByID(ctx context.Context, in *target.GetRequest) (*target.Empty, error) {
	logger.Info("Delete Targets")

	if in.ID == "" {
		err := errors.New("id must be provided")
		logger.Error(err)
		return &target.Empty{}, err
	}

	var fn func() error
	msg := "Deleting into targets Table for ID : "
	fn = func() error { return db.DeleteFromTargetDB(ctx, s.db, in.ID) }
	logger.Info(msg + in.ID)

	err := fn()

	logger.Info("done " + msg + in.ID)
	logger.With("id", in.ID).Info("data deleted")
	if err != nil {
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &target.Empty{}, err
}

// ListTargets implements target.ListTargets
func (s *server) ListTargets(_ *target.Empty, stream target.Target_ListTargetsServer) error {

	logger.Info("list targets")
	err := db.ListTargets(s.db, func(id, json string) error {
		return stream.Send(&target.TargetList{ID: id, Data: json})
	})

	if err != nil {
		return err
	}
	return nil
}
