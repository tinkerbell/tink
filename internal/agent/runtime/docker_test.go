package runtime_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tinkerbell/tink/internal/agent/runtime"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

func TestDocker(t *testing.T) {
	ctx := context.Background()
	rt, err := runtime.NewDocker()
	if err != nil {
		t.Fatal(err.Error())
	}

	// action := workflow.Action{
	// 	Name:  "foobar",
	// 	Image: "alpine",
	// 	Args:  []string{"sh", "-c", "echo -n Message > /tinkerbell/failure-message; echo -n Reason > /tinkerbell/failure-reason; exit 0"},
	// }

	action := workflow.Action{
		Name:  "foobar",
		Image: "alpine",
		Args:  []string{"sh", "-c", "trap \"\" SIGTERM; sleep 1m"},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		defer wait.Done()
		err = rt.Run(ctx, action)
		if err != nil {
			if i, ok := err.(interface{ FailureReason() string }); ok {
				fmt.Println(i.FailureReason())
			}
			t.Fatal(err)
		}
	}()

	wait.Wait()
}
