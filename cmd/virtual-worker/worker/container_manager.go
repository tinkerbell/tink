package worker

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	"github.com/tinkerbell/tink/internal/proto"
)

func getRandHexStr(r *rand.Rand, length int) string {
	alphabet := []byte("1234567890abcdef")
	resp := []byte{}
	for i := 0; i < length; i++ {
		resp = append(resp, alphabet[r.Intn(len(alphabet))])
	}
	return string(resp)
}

type fakeManager struct {
	// minimum milliseconds to sleep for faked Docker API calls
	sleepMinimum time.Duration
	// additional jitter milliseconds to sleep for faked Docker API calls
	sleepJitter time.Duration

	r      *rand.Rand
	logger logr.Logger
}

func (m *fakeManager) sleep() {
	jitter := time.Duration(m.r.Int31n(int32(m.sleepJitter.Milliseconds()))) * time.Millisecond
	time.Sleep(jitter + m.sleepMinimum)
}

// NewFakeContainerManager returns a fake worker.ContainerManager that will sleep for Docker API calls.
func NewFakeContainerManager(l logr.Logger, sleepMinimum, sleepJitter time.Duration) worker.ContainerManager {
	if sleepMinimum <= 0 {
		sleepMinimum = 1
	}
	if sleepJitter <= 0 {
		sleepJitter = 1
	}
	return &fakeManager{
		sleepMinimum: sleepMinimum,
		sleepJitter:  sleepJitter,
		logger:       l,
		// intentionally weak RNG. This is only for fake output
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *fakeManager) CreateContainer(_ context.Context, cmd []string, _ string, _ *proto.WorkflowAction, _, _ bool) (string, error) {
	m.logger.Info("creating container", "command", cmd)
	return getRandHexStr(m.r, 64), nil
}

func (m *fakeManager) StartContainer(_ context.Context, id string) error {
	m.logger.Info("starting container", "containerID", id)
	return nil
}

func (m *fakeManager) WaitForContainer(_ context.Context, id string) (proto.State, error) {
	m.logger.Info("waiting for container", "containerID", id)
	m.sleep()

	return proto.State_STATE_SUCCESS, nil
}

func (m *fakeManager) WaitForFailedContainer(_ context.Context, id string, failedActionStatus chan proto.State) {
	m.logger.Info("waiting for container", "containerID", id)
	m.sleep()
	failedActionStatus <- proto.State_STATE_SUCCESS
}

func (m *fakeManager) RemoveContainer(_ context.Context, id string) error {
	m.logger.Info("removing container", "containerID", id)
	return nil
}

func (m *fakeManager) PullImage(_ context.Context, image string) error {
	m.logger.Info("pulling image", "image", image)
	m.sleep()

	return nil
}
