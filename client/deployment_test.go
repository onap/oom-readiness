package client

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestIsDeploymentReady(t *testing.T) {
	tests := []struct {
		name      string
		expected  bool
		resources []runtime.Object
	}{
		{
			name:     "That deployment can be ready",
			expected: true,
			resources: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Name:       "cassandra-dc1-service",
						Namespace:  "onap",
						Generation: 1,
					},
					Status: appsv1.DeploymentStatus{
						Replicas:            3,
						UnavailableReplicas: 0,
						UpdatedReplicas:     0,
						ReadyReplicas:       3,
						ObservedGeneration:  1,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](3),
					},
				},
			},
		},
		{
			name:     "That deployment is not ready UnavailableReplicas",
			expected: false,
			resources: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Name:       "cassandra-dc1-service",
						Namespace:  "onap",
						Generation: 1,
					},
					Status: appsv1.DeploymentStatus{
						Replicas:            3,
						UnavailableReplicas: 1,
						UpdatedReplicas:     0,
						ReadyReplicas:       2,
						ObservedGeneration:  1,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](3),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			r := ReadinessClient{
				Client: fake.NewSimpleClientset(test.resources...),
			}

			ready := r.IsDeploymentReady("onap", "cassandra-dc1-service")

			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", true, ready)
			}
		})
	}
}
