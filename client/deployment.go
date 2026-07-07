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

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) IsDeploymentReady(namespace string, name string) bool {
	deployment, err := r.Client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error during get of deployment %s: %v", name, err)
		return false
	}
	if isDeploymentReady(*deployment) {
		log.Printf("Deployment %s is ready", name)
		return true
	} else {
		log.Printf("Deployment %s is NOT ready", name)
		return false
	}
}

func isDeploymentReady(dpl appsv1.Deployment) bool {
	return dpl.Status.UnavailableReplicas == 0 &&
		(dpl.Status.UpdatedReplicas == 0 || dpl.Status.UpdatedReplicas == *dpl.Spec.Replicas) &&
		dpl.Status.Replicas == *dpl.Spec.Replicas &&
		dpl.Status.ObservedGeneration == dpl.Generation
}
