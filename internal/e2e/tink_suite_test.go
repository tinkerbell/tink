//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/controller"
	"github.com/tinkerbell/tink/internal/grpcserver"
	"github.com/tinkerbell/tink/internal/server"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient  client.Client // You'll be using this client in your tests.
	testEnv    *envtest.Environment
	ctx        context.Context
	cancel     context.CancelFunc
	serverAddr string
	logger     logr.Logger
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	logger = zapr.NewLogger(zap.Must(zap.NewDevelopment()))

	// Installs CRDs into cluster
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	// Start the test cluster
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	cfg.Timeout = time.Second * 5 // Graceful shutdown of testenv for only 5s
	logger.Info("started test environment", "host", cfg.Host)

	// Add tink API to the client scheme
	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Create the K8s client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	errCh := make(chan error, 2)

	tinkServer, err := server.NewKubeBackedServerFromREST(logger, cfg, "default")
	Expect(err).To(Succeed())

	serverAddr, err = grpcserver.SetupGRPC(
		ctx,
		tinkServer,
		"127.0.0.1:0", // Randomly selected port
		errCh,
	)
	Expect(err).NotTo(HaveOccurred())
	logger.Info(fmt.Sprintf("HTTP server: %v", serverAddr))

	// Start the controller
	options := ctrl.Options{
		Logger: logger,
	}

	manager, err := controller.NewManager(cfg, options)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		err := manager.Start(ctx)
		Expect(err).To(BeNil())
	}()
})

var _ = AfterSuite(func() {
	By("Cancelling the context")

	By("stopping the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
