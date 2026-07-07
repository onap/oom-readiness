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

package main

import "testing"

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		podName     string
		jobName     string
		wantErr     bool
	}{
		{name: "no resource flag set is an error", wantErr: true},
		{name: "service name is enough", serviceName: "my-svc"},
		{name: "pod name is enough", podName: "my-pod"},
		{name: "job name is enough", jobName: "my-job"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateArgs(test.serviceName, test.podName, test.jobName)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateArgs(%q,%q,%q) error = %v, wantErr %t",
					test.serviceName, test.podName, test.jobName, err, test.wantErr)
			}
		})
	}
}
