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
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (r ReadinessClient) CheckServiceReadiness(namespace string, service_names []string) {

	timeout := 60 * time.Minute
	startTime := time.Now()
	for _, name := range service_names {
		// ready := r.IsJobComplete(job_name)
		for r.isServiceReady(namespace, name) != true {
			if time.Since(startTime) > timeout {
				slog.Warn("timed out waiting for to be ready", slog.String("job", name))
				os.Exit(1)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (r ReadinessClient) isServiceReady(namespace string, name string) bool {
	service, err := r.Client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error during get of service %s: %v", name, err)
	}

	var pods *v1.PodList
	if service.Spec.Selector != nil {
		pods, err = getPodsBySelectorLabels(service, r, namespace)
	} else {
		log.Print("No Selector found, check Endpoints")
		pods, err = getPodsByEndpoint(service, r, namespace)
	}

	if err != nil {
		log.Printf("Error during get of pods for service %s: %v", service.Name, err)
	}
	for _, pod := range pods.Items {
		log.Printf("Found pod %s selected by service %s", pod.Name, name)
		return r.IsPodReady(pod)
	}
	return false
}

func getPodsBySelectorLabels(service *v1.Service, r ReadinessClient, namespace string) (*v1.PodList, error) {
	labelSelector := metav1.LabelSelector{MatchLabels: service.Spec.Selector}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	return r.Client.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
}

func getPodsByEndpoint(service *v1.Service, r ReadinessClient, namespace string) (*v1.PodList, error) {
	endpoints, err := r.Client.CoreV1().Endpoints(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error while listing endpoints: %v", err)
	}
	for _, endpoint := range endpoints.Items {
		if strings.HasPrefix(endpoint.Name, service.Name) {
			addresses := endpoint.Subsets[0].Addresses
			if addresses != nil {
				pods := []v1.Pod{}
				for _, address := range addresses {
					name := address.TargetRef.Name
					log.Printf("Found pod %s selected by service %s", name, service.Name)
					pod, err := r.Client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
					if err != nil {
						log.Printf("Error while getting pod %s: %v", name, err)
					}
					pods = append(pods, *pod)
				}
				return &v1.PodList{Items: pods}, nil
			}
		}
	}
	log.Printf("No pods found that are selected by service %s", service.Name)
	return &v1.PodList{}, nil
}
