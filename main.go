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
	"os"

	readyclient "github.com/onap/readiness/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var namespace string
	var serviceName string
	var podName string
	var jobName string

	cli := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.StringVar(&serviceName, "service-name", "", "The name of the service to wait for")
	cli.StringVar(&podName, "pod-name", "", "The name of the pod to wait for")
	cli.StringVar(&jobName, "job-name", "", "The name of the job to wait for")
	cli.StringVar(&namespace, "namespace", "", "The Kubernetes namespace the resource is in")
	cli.Parse(os.Args[1:])

	client := kubernetesClient()
	readiness := readyclient.ReadinessClient{Client: client}
	if serviceName != "" {
		readiness.CheckServiceReadiness(namespace, []string{serviceName})
	}
	if jobName != "" {
		readiness.CheckJobReadiness(namespace, []string{jobName})
	}
	if podName != "" {
		readiness.CheckForPodReadiness(namespace, []string{podName})
	}
	if namespace == "" {
		namespace = os.Getenv("NAMESPACE")
	}
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
