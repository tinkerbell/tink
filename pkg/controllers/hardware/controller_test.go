package hardware

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var runtimescheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(runtimescheme)
	_ = v1alpha1.AddToScheme(runtimescheme)
}

func GetFakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(
		runtimescheme,
	).WithRuntimeObjects(
		&v1alpha1.Hardware{},
	)
}

func TestReconcile(t *testing.T) {
	cases := []struct {
		name         string
		seedHardware *v1alpha1.Hardware
		req          reconcile.Request
		want         reconcile.Result
		wantHw       *v1alpha1.Hardware
		wantErr      error
	}{
		{
			name: "ReconcileDiskData",
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hardware1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Metadata: &v1alpha1.HardwareMetadata{
						Instance: &v1alpha1.MetadataInstance{
							Storage: &v1alpha1.MetadataInstanceStorage{
								Disks: []*v1alpha1.MetadataInstanceStorageDisk{
									{
										Device: "/dev/sda",
									},
								},
							},
						},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "hardware1",
					Namespace: "default",
				},
			},
			want: reconcile.Result{},
			wantHw: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "hardware1",
					Namespace:       "default",
					ResourceVersion: "1000",
				},
				Spec: v1alpha1.HardwareSpec{
					Disks: []v1alpha1.Disk{
						{
							Device: "/dev/sda",
						},
					},
					Metadata: &v1alpha1.HardwareMetadata{
						Instance: &v1alpha1.MetadataInstance{
							Storage: &v1alpha1.MetadataInstanceStorage{
								Disks: []*v1alpha1.MetadataInstanceStorageDisk{
									{
										Device: "/dev/sda",
									},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tc := range cases {
		kc := GetFakeClientBuilder()
		if tc.seedHardware != nil {
			kc = kc.WithObjects(tc.seedHardware)
		}

		controller := &Controller{
			kubeClient: kc.Build(),
		}

		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := controller.Reconcile(context.Background(), tc.req)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
			if tc.want != got {
				t.Errorf("Got unexpected result. Wanted %v, got %v", tc.want, got)
				// Don't return, also check the modified object
			}
			hwflow := &v1alpha1.Hardware{}
			err := controller.kubeClient.Get(
				context.Background(),
				client.ObjectKey{Name: tc.wantHw.Name, Namespace: tc.wantHw.Namespace},
				hwflow)
			if err != nil {
				t.Errorf("Error finding desired hardware: %v", err)
				return
			}

			if diff := cmp.Diff(tc.wantHw, hwflow); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}
