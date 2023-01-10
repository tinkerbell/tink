package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHardwareTinkID(t *testing.T) {
	id := "d2c26e20-97e0-449c-b665-61efa7373f47"
	cases := []struct {
		name      string
		input     *Hardware
		want      string
		overwrite string
	}{
		{
			"Already set",
			&Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
					Annotations: map[string]string{
						HardwareIDAnnotation: id,
					},
				},
			},
			id,
			"",
		},
		{
			"nil annotations",
			&Hardware{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "debian",
					Namespace:   "default",
					Annotations: nil,
				},
			},
			"",
			"abc",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.input.TinkID() != tc.want {
				t.Errorf("Got unexpected ID: got %v, wanted %v", tc.input.TinkID(), tc.want)
			}

			tc.input.SetTinkID(tc.overwrite)

			if tc.input.TinkID() != tc.overwrite {
				t.Errorf("Got unexpected ID: got %v, wanted %v", tc.input.TinkID(), tc.overwrite)
			}
		})
	}
}
