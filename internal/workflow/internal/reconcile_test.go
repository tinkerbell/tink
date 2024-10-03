package internal_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/zerologr"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	tinkv1 "github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/ptr"
	. "github.com/tinkerbell/tink/internal/workflow/internal" //nolint:revive // Dot imports should not be used. Problem for another time though.
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	machineryruntimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileContext(t *testing.T) {
	ctx := context.Background()

	hw := newHardware(func(*tinkv1.Hardware) {})
	tmpl := newTemplate(func(t *tinkv1.Template) {
		t.Spec.Actions = []tinkv1.Action{
			{
				Name:  "action",
				Image: "image",
				Cmd:   ptr.String("{{ .Param.Foo }}"),
			},
		}
	})
	wrkflw := newWorkflow(func(w *tinkv1.Workflow) {
		w.Spec.HardwareRef = corev1.LocalObjectReference{Name: hw.Name}
		w.Spec.TemplateRef = corev1.LocalObjectReference{Name: tmpl.Name}
		w.Spec.TemplateParams = map[string]string{"Foo": "Bar"}
	})

	expectWrkflw := wrkflw.DeepCopy()
	expectWrkflw.Status.Actions = []tinkv1.ActionStatus{
		{
			Rendered: newAction(func(a *tinkv1.Action) {
				a.Name = "action"
				a.Image = "image"
				a.Cmd = ptr.String("Bar")
			}),
			State: "Pending",
			ID:    newActionID(),
		},
	}

	zl := zerolog.New(os.Stdout)
	logger := zerologr.New(&zl)

	scheme := runtime.NewScheme()
	machineryruntimeutil.Must(tinkv1.AddToScheme(scheme))

	clnt := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(hw, tmpl).
		Build()

	reconcileCtx := ReconciliationContext{
		Client:      clnt,
		Log:         logger,
		Workflow:    wrkflw,
		NewActionID: newActionID,
	}
	_, err := reconcileCtx.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expectWrkflw, wrkflw) {
		t.Fatal(cmp.Diff(expectWrkflw, wrkflw))
	}
}

func newWorkflow(fn func(*tinkv1.Workflow)) *tinkv1.Workflow {
	w := &tinkv1.Workflow{
		TypeMeta: v1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: tinkv1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "workflow",
		},
		Spec: tinkv1.WorkflowSpec{},
	}
	fn(w)
	return w
}

func newTemplate(fn func(*tinkv1.Template)) *tinkv1.Template {
	t := &tinkv1.Template{
		TypeMeta: v1.TypeMeta{
			Kind:       "Template",
			APIVersion: tinkv1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "template",
		},
	}
	fn(t)
	return t
}

func newHardware(fn func(*tinkv1.Hardware)) *tinkv1.Hardware {
	hw := &tinkv1.Hardware{
		TypeMeta: v1.TypeMeta{
			Kind:       "Hardware",
			APIVersion: tinkv1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "hardware",
		},
	}
	fn(hw)
	return hw
}

func newAction(fn func(*tinkv1.Action)) tinkv1.Action {
	a := tinkv1.Action{
		Args:    []string{},
		Env:     map[string]string{},
		Volumes: []tinkv1.Volume{},
	}
	fn(&a)
	return a
}

func newActionID() string {
	return "8659e46f-00ff-40e4-a19b-c8661ca81167"
}
