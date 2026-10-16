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
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	readyclient "github.com/onap/readiness/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// stringList collects every occurrence of a repeated flag.
type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var timeout int
	var namespace string
	var serviceNames stringList
	var podNames stringList
	var jobNames stringList

	cli := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.Var(&serviceNames, "service-name", "The name of a service to wait for (repeatable)")
	cli.Var(&podNames, "pod-name", "The name of a pod to wait for (repeatable)")
	cli.Var(&jobNames, "job-name", "The name of a job to wait for (repeatable)")
	cli.StringVar(&namespace, "namespace", "", "The Kubernetes namespace the resource is in")
	cli.IntVar(&timeout, "timeout", 10, "The time in minutes after which the check is failed")
	cli.Parse(os.Args[1:])

	if err := validateArgs(serviceNames, podNames, jobNames); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	timeoutDuration := time.Duration(timeout) * time.Minute

	client := kubernetesClient()
	readiness := readyclient.ReadinessClient{Client: client}
	if namespace == "" {
		namespace = os.Getenv("NAMESPACE")
	}
	if len(serviceNames) > 0 {
		readiness.CheckServiceReadiness(namespace, serviceNames, timeoutDuration)
	}
	if len(jobNames) > 0 {
		readiness.CheckJobReadiness(namespace, jobNames, timeoutDuration)
	}
	if len(podNames) > 0 {
		readiness.CheckPodReadiness(namespace, podNames, timeoutDuration)
	}
}

// validateArgs ensures at least one resource to wait for was requested.
// Without this guard a misconfigured init container (no/typo'd flag) would
// exit 0 immediately, letting the pod start before its dependency is ready.
func validateArgs(serviceNames []string, podNames []string, jobNames []string) error {
	if len(serviceNames) == 0 && len(podNames) == 0 && len(jobNames) == 0 {
		return errors.New("no resource to wait for: set at least one of --service-name, --pod-name or --job-name")
	}
	return nil
}

func kubernetesClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
