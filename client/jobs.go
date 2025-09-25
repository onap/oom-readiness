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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ReadinessClient) IsJobComplete(namespace string, job_name string) bool {
	log.Println("Checking readiness for job: ", job_name)

	job, err := r.Client.BatchV1().Jobs(namespace).Get(context.TODO(), job_name, metav1.GetOptions{})
	if err != nil {
		slog.Debug("Error occured during getting job: ", slog.Any("error", err))
		return false
	}
	succeeded := job.Status.Succeeded > 0
	if succeeded {
		log.Printf("Job '%s' succeeded", job_name)
	}
	return succeeded
}

func (r ReadinessClient) CheckJobReadiness(namespace string, job_names []string) {

	timeout := 60 * time.Minute
	startTime := time.Now()
	for _, job_name := range job_names {
		// ready := r.IsJobComplete(job_name)
		for r.IsJobComplete(namespace, job_name) != true {
			if time.Since(startTime) > timeout {
				slog.Warn("timed out waiting for to be ready", slog.String("job", job_name))
				os.Exit(1)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
