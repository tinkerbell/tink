//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	virtualworker "github.com/tinkerbell/tink/cmd/virtual-worker/worker"
	"github.com/tinkerbell/tink/internal/client"
	"github.com/tinkerbell/tink/internal/proto"
	googleproto "google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

func parseFile(filename string, obj interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, obj)
}

func createHardware(filename string) *v1alpha1.Hardware {
	ctx := context.Background()
	obj := &v1alpha1.Hardware{}
	err := parseFile(filename, obj)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Create(ctx, obj)
	Expect(err).NotTo(HaveOccurred())
	return obj
}

func createTemplate(filename string) *v1alpha1.Template {
	ctx := context.Background()
	obj := &v1alpha1.Template{}
	err := parseFile(filename, obj)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Create(ctx, obj)
	Expect(err).NotTo(HaveOccurred())
	return obj
}

func createWorkflow(filename string) *v1alpha1.Workflow {
	ctx := context.Background()
	obj := &v1alpha1.Workflow{}
	err := parseFile(filename, obj)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Create(ctx, obj)
	Expect(err).NotTo(HaveOccurred())
	return obj
}

var _ = Describe("Tink API", func() {
	Context("When a workflow is created", func() {
		It("01 - should complete a workflow", func() {
			ctx := context.Background()
			By("creating a hardware")
			hardware := createHardware(filepath.Join("./testdata/01", "hardware.yaml"))
			defer k8sClient.Delete(ctx, hardware)
			By("Creating a template")
			template := createTemplate(filepath.Join("./testdata/01", "template.yaml"))
			defer k8sClient.Delete(ctx, template)
			By("Creating a workflow")
			workflow := createWorkflow(filepath.Join("./testdata/01", "workflow.yaml"))
			defer k8sClient.Delete(ctx, workflow)

			By("Wait for the controller to update the workflow")
			timeout := time.Second * 2
			interval := time.Millisecond * 200
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, types.NamespacedName{Namespace: workflow.Namespace, Name: workflow.Name}, workflow)
				if err != nil {
					return "", err
				}
				return string(workflow.Status.State), nil
			}, timeout, interval).Should(Equal("STATE_PENDING"))

			By("Running a virtual worker")
			conn, err := client.NewClientConn(serverAddr, false)
			Expect(err).NotTo(HaveOccurred())
			rClient := proto.NewWorkflowServiceClient(conn)

			containerManager := virtualworker.NewFakeContainerManager(logger, time.Millisecond*100, time.Millisecond*200)
			logCapturer := virtualworker.NewEmptyLogCapturer()
			workerID := hardware.Spec.Interfaces[0].DHCP.MAC
			w := worker.NewWorker(
				workerID,
				rClient,
				containerManager,
				logCapturer,
				logger,
				worker.WithDataDir("./worker"),
				worker.WithMaxFileSize(1<<10),
				worker.WithRetries(time.Millisecond*500, 3))
			logger.Info("Created worker", "workerID", workerID)

			errChan := make(chan error)
			workerCtx, cancel := context.WithTimeout(ctx, time.Second*8)
			defer cancel()
			go func(errChan chan error) {
				err := w.ProcessWorkflowActions(workerCtx)
				errChan <- err
			}(errChan)
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, types.NamespacedName{Namespace: workflow.Namespace, Name: workflow.Name}, workflow)
				if err != nil {
					return "", err
				}
				return string(workflow.Status.State), nil
			}, 8*time.Second, 1*time.Second).Should(Equal("STATE_SUCCESS"))

			workerErr := <-errChan
			Expect(workerErr).To(BeNil())
		})

		It("02 - should return the correct workflow contexts", func() {
			By("creating hardware")
			hardware := createHardware(filepath.Join("./testdata/02", "hardware1.yaml"))
			defer k8sClient.Delete(ctx, hardware)

			By("Creating templates")
			template1 := createTemplate(filepath.Join("./testdata/02", "template1.yaml"))
			defer k8sClient.Delete(ctx, template1)
			template2 := createTemplate(filepath.Join("./testdata/02", "template2.yaml"))
			defer k8sClient.Delete(ctx, template2)
			template3 := createTemplate(filepath.Join("./testdata/02", "template3.yaml"))
			defer k8sClient.Delete(ctx, template3)

			By("Creating workflows")
			workflow1 := createWorkflow(filepath.Join("./testdata/02", "workflow1.yaml"))
			defer k8sClient.Delete(ctx, workflow1)
			workflow2 := createWorkflow(filepath.Join("./testdata/02", "workflow2.yaml"))
			defer k8sClient.Delete(ctx, workflow2)
			workflow3 := createWorkflow(filepath.Join("./testdata/02", "workflow3.yaml"))
			defer k8sClient.Delete(ctx, workflow3)

			By("Wait for the controller to update a workflow")
			timeout := time.Second * 2
			interval := time.Millisecond * 200
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, types.NamespacedName{Namespace: workflow3.Namespace, Name: workflow3.Name}, workflow3)
				if err != nil {
					return "", err
				}
				return string(workflow3.Status.State), nil
			}, timeout, interval).Should(Equal("STATE_PENDING"))

			By("Getting Workflow Contexts")
			conn, err := client.NewClientConn(serverAddr, false)
			Expect(err).NotTo(HaveOccurred())
			rClient := proto.NewWorkflowServiceClient(conn)
			workerID := hardware.Spec.Interfaces[0].DHCP.MAC
			res, err := rClient.GetWorkflowContexts(ctx, &proto.WorkflowContextRequest{WorkerId: workerID})
			Expect(err).NotTo(HaveOccurred())

			// expected workflow name to context mapping
			expectedWorkflows := map[string]*proto.WorkflowContext{
				"wf1": {
					WorkflowId:           "wf1",
					CurrentWorker:        "3c:ec:ef:4c:4f:54",
					CurrentTask:          "os-installation",
					CurrentAction:        "stream-image",
					CurrentActionIndex:   0,
					CurrentActionState:   proto.State_STATE_PENDING,
					TotalNumberOfActions: 3,
				},
				"wf3": {
					WorkflowId:           "wf3",
					CurrentWorker:        "3c:ec:ef:4c:4f:54",
					CurrentTask:          "task-1",
					CurrentAction:        "task-1-action-1",
					CurrentActionIndex:   0,
					CurrentActionState:   proto.State_STATE_PENDING,
					TotalNumberOfActions: 2,
				},
			}

			for got, err := res.Recv(); err == nil && got != nil; got, err = res.Recv() {
				want, ok := expectedWorkflows[got.WorkflowId]
				Expect(ok).To(BeTrue(), fmt.Sprintf("Didn't find expected context for %s", got.WorkflowId))
				if !ok {
					continue
				}
				if !googleproto.Equal(want, got) {
					fmt.Printf("Expected:\n\t%#v\nGot:\n\t%#v", want, got)
				}
				Expect(googleproto.Equal(want, got)).To(Equal(true), fmt.Sprintf("Didn't find expected context for %s", got.WorkflowId))

				// Remove the key from the map
				delete(expectedWorkflows, got.WorkflowId)
			}
			Expect(expectedWorkflows).To(BeEmpty(), "All expected workflows should be found")
		})
	})
})
