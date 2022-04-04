package convert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/internal/tests"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	prototemplate "github.com/tinkerbell/tink/protos/template"
	"google.golang.org/protobuf/testing/protocmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestTime is a static time that can be used for testing.
var TestTime = tests.NewFrozenTimeUnix(1637361794)

func TestTemplateCRDToProto(t *testing.T) {
	cases := []struct {
		name  string
		input *v1alpha1.Template
		want  *prototemplate.WorkflowTemplate
	}{
		{
			"Full Example",
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
					Annotations: map[string]string{
						"template.tinkerbell.org/id": "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
					},
					CreationTimestamp: *TestTime.MetaV1Now(),
					DeletionTimestamp: TestTime.MetaV1AfterSec(600),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: func() *string {
						resp := `version: "0.1"`
						return &resp
					}(),
				},
			},
			&prototemplate.WorkflowTemplate{
				Id:        "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				DeletedAt: TestTime.PbAfterSec(600),
				Data:      `version: "0.1"`,
			},
		},
		{
			"No DeletionTime",
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
					Annotations: map[string]string{
						"template.tinkerbell.org/id": "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
					},
					CreationTimestamp: *TestTime.MetaV1Now(),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: func() *string {
						resp := `version: "0.1"`
						return &resp
					}(),
				},
			},
			&prototemplate.WorkflowTemplate{
				Id:        "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				Data:      `version: "0.1"`,
			},
		},
		{
			"No Annotation or data",
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "template1",
					CreationTimestamp: *TestTime.MetaV1Now(),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: nil,
				},
			},
			&prototemplate.WorkflowTemplate{
				Id:        "",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				Data:      "",
			},
		},
		{
			"Empty",
			nil,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TemplateCRDToProto(tc.input)
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}

func TestTemplateProtoToCRD(t *testing.T) {
	cases := []struct {
		name  string
		input *prototemplate.WorkflowTemplate
		want  *v1alpha1.Template
	}{
		{
			"Full Example",
			&prototemplate.WorkflowTemplate{
				Id:        "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				DeletedAt: TestTime.PbAfterSec(600),
				Data:      `version: "0.1"`,
			},
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
					Annotations: map[string]string{
						"template.tinkerbell.org/id": "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
					},
					CreationTimestamp: *TestTime.MetaV1Now(),
					DeletionTimestamp: TestTime.MetaV1AfterSec(600),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: func() *string {
						resp := `version: "0.1"`
						return &resp
					}(),
				},
			},
		},
		{
			"No DeletionTime",
			&prototemplate.WorkflowTemplate{
				Id:        "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				Data:      `version: "0.1"`,
			},
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
					Annotations: map[string]string{
						"template.tinkerbell.org/id": "7d9031ee-18d4-4ba4-b934-c3a78a1330f6",
					},
					CreationTimestamp: *TestTime.MetaV1Now(),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: func() *string {
						resp := `version: "0.1"`
						return &resp
					}(),
				},
			},
		},
		{
			"No id or data",
			&prototemplate.WorkflowTemplate{
				Id:        "",
				Name:      "template1",
				CreatedAt: TestTime.PbNow(),
				Data:      "",
			},
			&v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "template1",
					Annotations: map[string]string{
						"template.tinkerbell.org/id": "",
					},
					CreationTimestamp: *TestTime.MetaV1Now(),
				},
				Spec: v1alpha1.TemplateSpec{
					Data: func(s string) *string { return &s }(""),
				},
			},
		},
		{
			"Empty",
			nil,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TemplateProtoToCRD(tc.input)
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}
