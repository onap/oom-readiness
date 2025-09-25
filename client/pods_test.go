package client

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
)

func TestIsPodReady(t *testing.T) {
	const name = "foo"
	const namespace = "onap"
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		OwnerReferences: []metav1.OwnerReference{
			{
				Name: "foo",
				Kind: "StatefulSet",
			},
		},
		Generation: 1,
	}
	tests := []struct {
		name      string
		expected  bool
		resources []runtime.Object
	}{
		{
			name:     "StatefulSet is ready",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: objectMeta,
				},
				&appsv1.StatefulSet{
					ObjectMeta: objectMeta,
					Spec: appsv1.StatefulSetSpec{
						Replicas: pointer.Int32(3),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:           3,
						ReadyReplicas:      3,
						ObservedGeneration: 1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readiness := &ReadinessClient{
				Client: fake.NewSimpleClientset(test.resources...),
			}

			ready := readiness.IsPodReady(*test.resources[0].(*corev1.Pod))
			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", test.expected, ready)
			}
		})
	}
}
