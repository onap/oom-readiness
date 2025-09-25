package client

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIsStatefulSetReady(t *testing.T) {
	const name = "cassandra-dc1-default-sts"
	const namespace = "onap"

	tests := []struct {
		name     string
		expected bool
		meta     metav1.ObjectMeta
		spec     appsv1.StatefulSetSpec
		status   appsv1.StatefulSetStatus
	}{
		{
			name:     "StatefulSet is ready when all replicas are up",
			expected: true,
			meta: metav1.ObjectMeta{
				Generation: 1,
			},
			spec: appsv1.StatefulSetSpec{
				Replicas: pointer.Int32(3),
			},
			status: appsv1.StatefulSetStatus{
				Replicas:           3,
				ReadyReplicas:      3,
				ObservedGeneration: 1,
			},
		},
		{
			name:     "StatefulSet is not ready when replicas are not ready",
			expected: false,
			meta: metav1.ObjectMeta{
				Generation: 1,
			},
			spec: appsv1.StatefulSetSpec{
				Replicas: pointer.Int32(3),
			},
			status: appsv1.StatefulSetStatus{
				Replicas:           3,
				ReadyReplicas:      2,
				ObservedGeneration: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resources := []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:       name,
						Namespace:  namespace,
						Generation: test.meta.Generation,
					},
					Spec:   test.spec,
					Status: test.status,
				},
			}
			readiness := &ReadinessClient{
				Client: fake.NewSimpleClientset(resources...),
			}

			ready := readiness.IsStatefulSetReady(namespace, name)
			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", test.expected, ready)
			}
		})
	}
}
