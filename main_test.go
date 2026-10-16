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

import (
	"flag"
	"reflect"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name         string
		serviceNames []string
		podNames     []string
		jobNames     []string
		wantErr      bool
	}{
		{name: "no resource flag set is an error", wantErr: true},
		{name: "service name is enough", serviceNames: []string{"my-svc"}},
		{name: "pod name is enough", podNames: []string{"my-pod"}},
		{name: "job name is enough", jobNames: []string{"my-job"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateArgs(test.serviceNames, test.podNames, test.jobNames)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateArgs(%q,%q,%q) error = %v, wantErr %t",
					test.serviceNames, test.podNames, test.jobNames, err, test.wantErr)
			}
		})
	}
}

func TestStringListKeepsEveryOccurrence(t *testing.T) {
	var services stringList
	cli := flag.NewFlagSet("readiness", flag.ContinueOnError)
	cli.Var(&services, "service-name", "")

	err := cli.Parse([]string{"--service-name", "sdc-be", "--service-name", "cds-blueprints-processor-http"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := stringList{"sdc-be", "cds-blueprints-processor-http"}
	if !reflect.DeepEqual(services, want) {
		t.Fatalf("services = %q, want %q", services, want)
	}
}
