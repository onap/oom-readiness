// -
//   ========================LICENSE_START=================================
//   Copyright (C) 2025: Deutsche Telekom
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   SPDX-License-Identifier: Apache-2.0
//   ========================LICENSE_END===================================

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
	objectMeta := v1.ObjectMeta{
		Name:       "cassandra-dc1-service",
		Namespace:  "onap",
		Generation: 1,
	}
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
					ObjectMeta: objectMeta,
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
					ObjectMeta: objectMeta,
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
		{
			name:     "That deployment is not ready UpdatedReplicas",
			expected: false,
			resources: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: objectMeta,
					Status: appsv1.DeploymentStatus{
						Replicas:            3,
						UnavailableReplicas: 0,
						UpdatedReplicas:     1,
						ReadyReplicas:       0,
						ObservedGeneration:  1,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](3),
					},
				},
			},
		},
		{
			name:     "That deployment is not ready SpecStatusReplicas",
			expected: false,
			resources: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: objectMeta,
					Status: appsv1.DeploymentStatus{
						Replicas:            2,
						UnavailableReplicas: 0,
						UpdatedReplicas:     0,
						ReadyReplicas:       0,
						ObservedGeneration:  1,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](3),
					},
				},
			},
		},
		{
			name:     "That deployment is not ready Generation",
			expected: false,
			resources: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Name:       "cassandra-dc1-service",
						Namespace:  "onap",
						Generation: 2,
					},
					Status: appsv1.DeploymentStatus{
						Replicas:            3,
						UnavailableReplicas: 0,
						UpdatedReplicas:     0,
						ReadyReplicas:       0,
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
