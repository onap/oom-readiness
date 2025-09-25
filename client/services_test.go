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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestIsServiceReady(t *testing.T) {
	selectorLabels := map[string]string{
		"cassandra.datastax.com/cluster":    "cassandra",
		"cassandra.datastax.com/datacenter": "dc1",
	}

	tests := []struct {
		name        string
		serviceSpec corev1.ServiceSpec
		endpoints   corev1.EndpointsList
	}{
		{
			name: "Service with Selector",
			serviceSpec: corev1.ServiceSpec{
				Selector: selectorLabels,
			},
		},
		{
			name: "Service without Selector, but with Endpoints",
			serviceSpec: corev1.ServiceSpec{
				Selector: nil,
			},
			endpoints: corev1.EndpointsList{
				Items: []corev1.Endpoints{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cassandra-dc1-service",
							Namespace: "onap",
						},
						Subsets: []corev1.EndpointSubset{
							{
								Addresses: []corev1.EndpointAddress{
									{
										TargetRef: &corev1.ObjectReference{
											Name: "cassandra-dc1-default-sts-0",
										},
									},
									{
										TargetRef: &corev1.ObjectReference{
											Name: "cassandra-dc1-default-sts-1",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resources := []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cassandra-dc1-service",
						Namespace: "onap",
					},
					Spec: test.serviceSpec,
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cassandra-dc1-default-sts-0",
						Namespace: "onap",
						Labels:    selectorLabels,
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "StatefulSet",
								Name: "cassandra-dc1-default-sts",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cassandra-dc1-default-sts-1",
						Namespace: "onap",
						Labels:    selectorLabels,
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "StatefulSet",
								Name: "cassandra-dc1-default-sts",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "cassandra-dc1-default-sts",
						Namespace:  "onap",
						Generation: 1,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](3),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:           3,
						ReadyReplicas:      3,
						ObservedGeneration: 1,
					},
				},
				&test.endpoints,
			}

			r := ReadinessClient{
				Client: fake.NewSimpleClientset(resources...),
			}

			ready := r.isServiceReady("onap", "cassandra-dc1-service")
			if ready != true {
				t.Fatalf("expected ready to be %t, but was %t", true, ready)
			}
		})
	}
}
