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

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIsJobComplete(t *testing.T) {
	testcases := []struct {
		name     string
		expected bool
		job_name string
		jobs     []runtime.Object
	}{
		{
			name:     "Job_is_ready",
			expected: true,
			job_name: "pod1",
			jobs: []runtime.Object{
				&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "namespace1",
						Labels: map[string]string{
							"label1": "value1",
						},
					},
					Status: batchv1.JobStatus{
						Succeeded: 1,
					},
				},
			},
		},
		{
			name:     "Errors return ready=false",
			expected: false,
			job_name: "unknownjob",
			jobs: []runtime.Object{
				&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
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
			ready := readiness.IsJobComplete("namespace1", test.job_name)
			if ready != test.expected {
				t.Fatalf("expected ready to be %t, but was %t", test.expected, ready)
			}
		})
	}
}

func TestCheckJobReadiness(t *testing.T) {
	testcases := []struct {
		name      string
		job_names []string
		jobs      []runtime.Object
	}{
		{
			name:      "That method runs until IsJobComplete returns true",
			job_names: []string{"someJob"},
			jobs: []runtime.Object{
				&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "someJob",
						Namespace: "namespace1",
					},
					Status: batchv1.JobStatus{
						Succeeded: 1,
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
			readiness.CheckJobReadiness("namespace1", test.job_names, time.Duration(10))
		})
	}
}
