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
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) IsPodReady(pod corev1.Pod) bool {

	ownerReference := pod.ObjectMeta.OwnerReferences[0]
	switch resource := ownerReference.Kind; resource {
	case "StatefulSet":
		return r.IsStatefulSetReady(pod.Namespace, ownerReference.Name)
	case "ReplicaSet":
		deploymentName := getDeploymentFromReplicaSet(r, pod.Namespace, ownerReference.Name)
		if deploymentName == "" {
			return false
		}
		return r.IsDeploymentReady(pod.Namespace, deploymentName)
	case "Job":
		return r.IsJobComplete(pod.Namespace, pod.Name)
	case "DaemonSet":
		return r.isDaemonSetReady(pod.Namespace, pod.Name)
	}

	return true
}

func (r ReadinessClient) CheckPodReadiness(namespace string, names []string, timeout time.Duration) {
	for _, name := range names {
		podsWithName := r.getPodsByName(namespace, name)
		for _, pod := range podsWithName {
			go waitForPod(r, pod, timeout)
		}
	}
}

func waitForPod(r ReadinessClient, pod corev1.Pod, timeout time.Duration) {
	startTime := time.Now()
	for r.IsPodReady(pod) != true {
		if time.Since(startTime) > timeout*time.Minute {
			log.Printf("Timed out waiting for pod %s to be ready", pod.Name)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}
}

// pods have a partially dynamic name, i.e onap-aai-resources-f7f6d5cf4-mqzv7 and
// not always a fixed label (like app.kubernetes.io/name=aai-resources)
// therefore it is necessary to fetch the whole list of pods and manually filter
func (r ReadinessClient) getPodsByName(namespace string, name string) []corev1.Pod {
	var pods *corev1.PodList
	var err error
	_continue := ""
	result := []corev1.Pod{}
	for true {
		// _continue is the pagination index. In the first run of this loop, it is not defined yet
		if _continue == "" {
			pods, err = r.Client.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{Limit: 300})
		} else {
			pods, err = r.Client.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{Limit: 300, Continue: _continue})
		}
		if err != nil {
			log.Printf("Failed to list pods: %v", err)
		}

		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, name) {
				result = append(result, pod)
				break
			}
		}
		_continue = pods.Continue
	}
	return result
}

func getDeploymentFromReplicaSet(r ReadinessClient, namespace string, name string) string {
	replicaSet, err := r.Client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		log.Printf("Error during get of ReplicaSet %s: %v", name, err)
		return ""
	}
	return replicaSet.ObjectMeta.OwnerReferences[0].Name
}
