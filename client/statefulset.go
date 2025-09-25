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

// Package main is the entry point for the policy-opa-pdp service.
// This package initializes the HTTP server, Kafka consumer and producer, and handles
// the overall service lifecycle including graceful shutdown

package client

import (
	"context"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) IsStatefulSetReady(namespace string, name string) bool {
	sts, err := r.Client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		log.Printf("Error while get for StatefulSet %s: %v", sts.Name, err)
	}
	if isReady(sts) {
		log.Printf("StatefulSet %s is ready", sts.Name)
		return true
	} else {
		log.Printf("StatefulSet %s is NOT ready", sts.Name)
		return false
	}
}

func isReady(sts *appsv1.StatefulSet) bool {
	return sts.Status.Replicas == *sts.Spec.Replicas &&
		sts.Status.ReadyReplicas == *sts.Spec.Replicas &&
		sts.Status.ObservedGeneration == sts.ObjectMeta.Generation
}
