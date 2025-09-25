package client

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIsDaemonSetReady(t *testing.T) {
	testcases := []struct {
		name     string
		expected bool
		job_name string
		jobs     []runtime.Object
	}{
		{
			name:     "DaemonSet is ready",
			expected: true,
			job_name: "pod1",
			jobs: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: v1.ObjectMeta{
						Name:      "pod1",
						Namespace: "namespace1",
						Labels: map[string]string{
							"label1": "value1",
						},
					},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 3,
						NumberReady:            3,
					},
				},
			},
		},
		{
			name:     "DaemonSet is NOT ready",
			expected: false,
			job_name: "pod1",
			jobs: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: v1.ObjectMeta{
						Name:      "pod1",
						Namespace: "namespace1",
						Labels: map[string]string{
							"label1": "value1",
						},
					},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 3,
						NumberReady:            2,
					},
				},
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			readiness := &ReadinessClient{
				Client: fake.NewSimpleClientset(test.jobs...),
			}
			ready := readiness.isDaemonSetReady("namespace1", test.job_name)
			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", test.expected, ready)
			}
		})
	}
}
