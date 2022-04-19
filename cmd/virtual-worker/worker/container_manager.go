package worker

import (
	"context"
	"math/rand"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	pb "github.com/tinkerbell/tink/protos/workflow"
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
	logger log.Logger
}

func (m *fakeManager) sleep() {
	jitter := time.Duration(m.r.Int31n(int32(m.sleepJitter.Milliseconds()))) * time.Millisecond
	time.Sleep(jitter + m.sleepMinimum)
}

// NewFakeContainerManager returns a fake worker.ContainerManager that will sleep for Docker API calls.
func NewFakeContainerManager(l log.Logger, sleepMinimum, sleepJitter time.Duration) worker.ContainerManager {
	return &fakeManager{
		sleepMinimum: sleepMinimum,
		sleepJitter:  sleepJitter,
		logger:       l,
		// intentionally weak RNG. This is only for fake output
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *fakeManager) CreateContainer(_ context.Context, cmd []string, _ string, _ *pb.WorkflowAction, _, _ bool) (string, error) {
	m.logger.With("command", cmd).Info("creating container")
	return getRandHexStr(m.r, 64), nil
}

func (m *fakeManager) StartContainer(_ context.Context, id string) error {
	m.logger.With("containerID", id).Debug("starting container")
	return nil
}

func (m *fakeManager) WaitForContainer(_ context.Context, id string) (pb.State, error) {
	m.logger.With("containerID", id).Info("waiting for container")
	m.sleep()

	return pb.State_STATE_SUCCESS, nil
}

func (m *fakeManager) WaitForFailedContainer(_ context.Context, id string, failedActionStatus chan pb.State) {
	m.logger.With("containerID", id).Info("waiting for container")
	m.sleep()
	failedActionStatus <- pb.State_STATE_SUCCESS
}

func (m *fakeManager) RemoveContainer(_ context.Context, id string) error {
	m.logger.With("containerID", id).Info("removing container")
	return nil
}

func (m *fakeManager) PullImage(_ context.Context, image string) error {
	m.logger.With("image", image).Info("pulling image")
	m.sleep()

	return nil
}
