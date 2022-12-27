package e2e_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/internal/grpcserver"
	"github.com/tinkerbell/tink/internal/server"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/pkg/controllers"
	wfctrl "github.com/tinkerbell/tink/pkg/controllers/workflow"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient  client.Client // You'll be using this client in your tests.
	testEnv    *envtest.Environment
	ctx        context.Context
	cancel     context.CancelFunc
	serverAddr string
	logger     log.Logger
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	// Create Test Tink API gRPC Server
	logger, err = log.Init("github.com/tinkerbell/tink/tests")
	Expect(err).NotTo(HaveOccurred())

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
	logger.With("host", cfg.Host).Info("started test environment")

	// Add tink API to the client scheme
	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Create the K8s client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// database, err := db.NewK8sDatabaseFromREST(cfg, logger)
	// Expect(err).NotTo(HaveOccurred())
	errCh := make(chan error, 2)

	tinkServer := server.NewKubeBackedServerFromREST(logger,
		cfg,
		"default",
	)
	serverAddr, err = grpcserver.SetupGRPC(
		ctx,
		tinkServer,
		"127.0.0.1:0", // Randomly selected port
		errCh,
	)
	Expect(err).NotTo(HaveOccurred())
	logger.Info("HTTP server: ", fmt.Sprintf("%+v", serverAddr))

	// Start the controller
	options := controllers.GetControllerOptions()
	options.LeaderElectionNamespace = "default"
	manager, err := controllers.NewManager(cfg, options)
	Expect(err).NotTo(HaveOccurred())
	go func() {
		err := manager.RegisterControllers(ctx, wfctrl.NewController(manager.GetClient())).Start(ctx)
		Expect(err).To(BeNil())
	}()
})

var _ = AfterSuite(func() {
	By("Cancelling the context")
	cancel()
	By("stopping the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
