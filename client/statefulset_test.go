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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

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
				Replicas: ptr.To[int32](3),
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
				Replicas: ptr.To[int32](3),
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

// A transient API error (e.g. the StatefulSet not existing yet) must report
// "not ready" rather than dereferencing the nil Spec.Replicas of a zero-value
// object.
func TestIsStatefulSetReadyReturnsFalseOnError(t *testing.T) {
	readiness := ReadinessClient{Client: fake.NewSimpleClientset()}

	if readiness.IsStatefulSetReady("onap", "does-not-exist") {
		t.Fatal("expected not-ready when the StatefulSet cannot be fetched")
	}
}
