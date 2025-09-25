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
	corev1 "k8s.io/api/core/v1"
)

func (r ReadinessClient) IsPodReady(pod corev1.Pod) bool {
	// _ready := false
	ownerReference := pod.ObjectMeta.OwnerReferences[0]
	switch resource := ownerReference.Kind; resource {
	case "StatefulSet":
		return r.IsStatefulSetReady(pod.Namespace, ownerReference.Name)
	}

	return true
}
