package transport_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/transport"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	workflowproto "github.com/tinkerbell/tink/internal/proto/workflow/v2"
	"google.golang.org/grpc"
)

func TestGRPC(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	type streamResponse struct {
		Workflow *workflowproto.GetWorkflowsResponse
		Error    error
	}
	responses := make(chan streamResponse, 2)
	responses <- streamResponse{
		Workflow: &workflowproto.GetWorkflowsResponse{
			Cmd: &workflowproto.GetWorkflowsResponse_StartWorkflow_{
				StartWorkflow: &workflowproto.GetWorkflowsResponse_StartWorkflow{
					Workflow: &workflowproto.Workflow{},
				},
			},
		},
	}
	responses <- streamResponse{
		Error: io.EOF,
	}

	stream := &workflowproto.WorkflowService_GetWorkflowsClientMock{
		RecvFunc: func() (*workflowproto.GetWorkflowsResponse, error) {
			r, ok := <-responses
			if !ok {
				return nil, io.EOF
			}
			return r.Workflow, r.Error
		},
		ContextFunc: context.Background,
	}
	client := &workflowproto.WorkflowServiceClientMock{
		GetWorkflowsFunc: func(ctx context.Context, in *workflowproto.GetWorkflowsRequest, opts ...grpc.CallOption) (workflowproto.WorkflowService_GetWorkflowsClient, error) {
			return stream, nil
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	handler := &transport.WorkflowHandlerMock{
		HandleWorkflowFunc: func(contextMoqParam context.Context, workflow workflow.Workflow, recorder event.Recorder) {
			defer wg.Done()
			close(responses)
		},
	}

	g := transport.NewGRPC(zerologr.New(&logger), client)

	err := g.Start(context.Background(), "id", handler)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
}
