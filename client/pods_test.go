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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestIsPodReady(t *testing.T) {
	const namespace = "onap"
	tests := []struct {
		name      string
		expected  bool
		resources []runtime.Object
	}{
		{
			name:     "Pod owned by a ready StatefulSet is ready",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cassandra-0",
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "StatefulSet", Name: "cassandra"},
						},
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "cassandra", Namespace: namespace, Generation: 1},
					Spec:       appsv1.StatefulSetSpec{Replicas: ptr.To[int32](3)},
					Status:     appsv1.StatefulSetStatus{Replicas: 3, ReadyReplicas: 3, ObservedGeneration: 1},
				},
			},
		},
		{
			// Owner resolution must use the owner's name, not the pod's (suffixed) name.
			name:     "Pod owned by a completed Job is ready",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "migration-abcde",
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "Job", Name: "migration"},
						},
					},
				},
				&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{Name: "migration", Namespace: namespace},
					Status:     batchv1.JobStatus{Succeeded: 1},
				},
			},
		},
		{
			// Owner resolution must use the owner's name, not the pod's (suffixed) name.
			name:     "Pod owned by a ready DaemonSet is ready",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "log-agent-abcde",
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "DaemonSet", Name: "log-agent"},
						},
					},
				},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "log-agent", Namespace: namespace},
					Status:     appsv1.DaemonSetStatus{DesiredNumberScheduled: 2, NumberReady: 2},
				},
			},
		},
		{
			name:     "Pod owned by a ReplicaSet resolves to its ready Deployment",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "portal-7f6d5cf4-mqzv7",
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "ReplicaSet", Name: "portal-7f6d5cf4"},
						},
					},
				},
				&appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "portal-7f6d5cf4",
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "Deployment", Name: "portal"},
						},
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "portal", Namespace: namespace, Generation: 1},
					Spec:       appsv1.DeploymentSpec{Replicas: ptr.To[int32](2)},
					Status: appsv1.DeploymentStatus{
						Replicas:           2,
						ReadyReplicas:      2,
						UpdatedReplicas:    2,
						ObservedGeneration: 1,
					},
				},
			},
		},
		{
			// A pod without an owning controller must not panic on OwnerReferences[0].
			name:     "Pod without an owner does not panic",
			expected: true,
			resources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "standalone",
						Namespace: namespace,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readiness := ReadinessClient{
				Client: fake.NewSimpleClientset(test.resources...),
			}

			ready := readiness.IsPodReady(*test.resources[0].(*corev1.Pod))
			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", test.expected, ready)
			}
		})
	}
}

// TestGetPodsByName guards against the pagination loop never terminating: once the
// final page is reached the Continue token is empty, which must end the loop rather
// than restart it from the first page.
func TestGetPodsByName(t *testing.T) {
	const namespace = "onap"
	r := ReadinessClient{
		Client: fake.NewSimpleClientset(
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "onap-aai-resources-f7f6d5cf4-mqzv7", Namespace: namespace}},
			&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "onap-sdc-fe-abc123", Namespace: namespace}},
		),
	}

	done := make(chan []corev1.Pod, 1)
	go func() {
		done <- r.getPodsByName(namespace, "onap-aai-resources")
	}()

	select {
	case pods := <-done:
		if len(pods) != 1 {
			t.Fatalf("expected 1 matching pod, got %d", len(pods))
		}
		if pods[0].Name != "onap-aai-resources-f7f6d5cf4-mqzv7" {
			t.Fatalf("unexpected pod matched: %q", pods[0].Name)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("getPodsByName did not return within 5s — pagination loop never terminates")
	}
}

// CheckPodReadiness spawns a goroutine per matching pod; it must block until they
// have all observed readiness rather than returning while they are still running.
func TestCheckPodReadinessWaitsForPods(t *testing.T) {
	const namespace = "onap"
	r := ReadinessClient{
		Client: fake.NewSimpleClientset(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cassandra-0",
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						{Kind: "StatefulSet", Name: "cassandra"},
					},
				},
			},
			&appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "cassandra", Namespace: namespace, Generation: 1},
				Spec:       appsv1.StatefulSetSpec{Replicas: ptr.To[int32](1)},
				Status:     appsv1.StatefulSetStatus{Replicas: 1, ReadyReplicas: 1, ObservedGeneration: 1},
			},
		),
	}

	done := make(chan struct{})
	go func() {
		r.CheckPodReadiness(namespace, []string{"cassandra"}, 5*time.Second)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("CheckPodReadiness did not return for a ready pod within 5s")
	}
}
