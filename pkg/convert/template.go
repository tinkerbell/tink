package convert

import (
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	prototemplate "github.com/tinkerbell/tink/protos/template"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func metav1ToTimestamppb(in *metav1.Time) *timestamppb.Timestamp {
	if in == nil {
		return nil
	}
	return timestamppb.New(in.Time)
}

// TemplateCRDToProto converts a K8s Template to a tinkerbell WorkflowTemplate.
func TemplateCRDToProto(t *v1alpha1.Template) *prototemplate.WorkflowTemplate {
	if t == nil {
		return nil
	}
	data := ""
	if t.Spec.Data != nil {
		data = *t.Spec.Data
	}
	return &prototemplate.WorkflowTemplate{
		Id:        t.TinkID(),
		Name:      t.Name,
		CreatedAt: timestamppb.New(t.CreationTimestamp.Time),
		DeletedAt: metav1ToTimestamppb(t.DeletionTimestamp),
		Data:      data,
	}
}

// TemplateProtoToCRD converts a tinkerbell WorkflowTemplate to a K8s Template.
func TemplateProtoToCRD(t *prototemplate.WorkflowTemplate) *v1alpha1.Template {
	if t == nil {
		return nil
	}
	return &v1alpha1.Template{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Template",
			APIVersion: "tinkerbell.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.Name,
			Annotations: map[string]string{
				"template.tinkerbell.org/id": t.Id,
			},
			CreationTimestamp: metav1.NewTime(t.CreatedAt.AsTime()),
			DeletionTimestamp: func() *metav1.Time {
				if t.DeletedAt != nil {
					resp := metav1.NewTime(t.DeletedAt.AsTime())
					return &resp
				}
				return nil
			}(),
		},
		Spec: v1alpha1.TemplateSpec{
			Data: &t.Data,
		},
	}
}
