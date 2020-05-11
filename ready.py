#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright © 2020 Orange
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Kubernetes readiness check.

Checks if a container is ready or if a job is finished.
The check is done according to the name of the container, not the name of
its parent (Job, Deployment, StatefulSet, DaemonSet).
"""

import getopt
import logging
import os
import sys
import time
import random

from kubernetes import client
from kubernetes.client.rest import ApiException

# extract env variables.
namespace = os.environ['NAMESPACE']
cert = os.environ['CERT']
host = os.environ['KUBERNETES_SERVICE_HOST']
token_path = os.environ['TOKEN']

with open(token_path, 'r') as token_file:
    token = token_file.read().replace('\n', '')

# setup logging
log = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
handler.setLevel(logging.INFO)
log.addHandler(handler)
log.setLevel(logging.INFO)

configuration = client.Configuration()
configuration.host = "https://" + host
configuration.ssl_ca_cert = cert
configuration.api_key['authorization'] = token
configuration.api_key_prefix['authorization'] = 'Bearer'
coreV1Api = client.CoreV1Api(client.ApiClient(configuration))
api_instance = client.ExtensionsV1beta1Api(client.ApiClient(configuration))
api = client.AppsV1beta1Api(client.ApiClient(configuration))
batchV1Api = client.BatchV1Api(client.ApiClient(configuration))


def is_job_complete(job_name):
    """
    Check if Job is complete.

    Args:
        job_name (str): the name of the Job.

    Returns:
        True if job is complete, false otherwise
    """
    complete = False
    log.info("Checking if %s is complete", job_name)
    try:
        response = batchV1Api.read_namespaced_job_status(job_name, namespace)
        if response.status.succeeded == 1:
            job_status_type = response.status.conditions[0].type
            if job_status_type == "Complete":
                complete = True
                log.info("%s is complete", job_name)
            else:
                log.info("%s is NOT complete", job_name)
        else:
            log.info("%s has not succeeded yet", job_name)
    except ApiException as exc:
        log.error("Exception when calling read_namespaced_job_status: %s\n",
                  exc)
    return complete


def wait_for_statefulset_complete(statefulset_name):
    """
    Check if StatefulSet is running.

    Args:
        statefulset_name (str): the name of the StatefulSet.

    Returns:
        True if StatefulSet is running, false otherwise
    """
    complete = False
    try:
        response = api.read_namespaced_stateful_set(statefulset_name,
                                                    namespace)
        status = response.status
        if (status.replicas == response.spec.replicas and
                status.ready_replicas == response.spec.replicas and
                status.observed_generation == response.metadata.generation):
            log.info("Statefulset %s is ready", statefulset_name)
            complete = True
        else:
            log.info("Statefulset %s is NOT ready", statefulset_name)
    except ApiException as exc:
        log.error("Exception when waiting for Statefulset status: %s\n", exc)
    return complete


def wait_for_deployment_complete(deployment_name):
    """
    Check if Deployment is running.

    Args:
        deployment_name (str): the name of the Deployment.

    Returns:
        True if Deployment is running, false otherwise
    """
    complete = False
    try:
        response = api.read_namespaced_deployment(deployment_name, namespace)
        status = response.status
        if (status.unavailable_replicas is None and
                (status.updated_replicas is None or
                 status.updated_replicas == response.spec.replicas) and
                status.replicas == response.spec.replicas and
                status.ready_replicas == response.spec.replicas and
                status.observed_generation == response.metadata.generation):
            log.info("Deployment %s is ready", deployment_name)
            complete = True
        else:
            log.info("Deployment %s is NOT ready", deployment_name)
    except ApiException as exc:
        log.error("Exception when waiting for deployment status: %s\n", exc)
    return complete


def wait_for_daemonset_complete(daemonset_name):
    """
    Check if DaemonSet is running.

    Args:
        daemonset_name (str): the name of the DaemonSet.

    Returns:
        True if DaemonSet is running, false otherwise
    """
    complete = False
    try:
        response = api_instance.read_namespaced_daemon_set(
            daemonset_name, namespace)
        status = response.status
        if status.desired_number_scheduled == status.number_ready:
            log.info("DaemonSet: %s/%s nodes ready --> %s is ready",
                     status.number_ready, status.desired_number_scheduled,
                     daemonset_name)
            complete = True
        else:
            log.info("DaemonSet: %s/%s nodes ready --> %s is NOT ready",
                     status.number_ready, status.desired_number_scheduled,
                     daemonset_name)
    except ApiException as exc:
        log.error("Exception when waiting for DaemonSet status: %s\n", exc)
    return complete


def is_ready(container_name):
    """
    Check if a container is ready.

    For a container owned by a Job, it means the Job is complete.
    Otherwise, it means the parent (Deployment, StatefulSet, DaemonSet) is
    running with the right number of replicas

    Args:
        container_name (str): the name of the container.

    Returns:
        True if container is ready, false otherwise
    """
    ready = False
    log.info("Checking if %s is ready", container_name)
    try:
        response = coreV1Api.list_namespaced_pod(namespace=namespace,
                                                 watch=False)
        for item in response.items:
            # container_statuses can be None, which is non-iterable.
            if item.status.container_statuses is None:
                continue
            for container in item.status.container_statuses:
                if container.name == container_name:
                    name = read_name(item)
                    if item.metadata.owner_references[0].kind == "StatefulSet":
                        ready = wait_for_statefulset_complete(name)
                    elif item.metadata.owner_references[0].kind == "ReplicaSet":
                        deployment_name = get_deployment_name(name)
                        ready = wait_for_deployment_complete(deployment_name)
                    elif item.metadata.owner_references[0].kind == "Job":
                        ready = is_job_complete(name)
                    elif item.metadata.owner_references[0].kind == "DaemonSet":
                        ready = wait_for_daemonset_complete(
                            item.metadata.owner_references[0].name)
                    return ready
    except ApiException as exc:
        log.error("Exception when calling list_namespaced_pod: %s\n", exc)
    return ready


def read_name(item):
    """
    Return the name of the owner's item.

    Args:
        item (str): the item.

    Returns:
        the name of first owner's item
    """
    return item.metadata.owner_references[0].name


def get_deployment_name(replicaset):
    """
    Return the name of the Deployment owning the ReplicatSet.

    Args:
        replicaset (str): the ReplicatSet.

    Returns:
        the name of the Deployment owning the ReplicatSet
    """
    api_response = api_instance.read_namespaced_replica_set_status(replicaset,
                                                                   namespace)
    deployment_name = read_name(api_response)
    return deployment_name


DEF_TIMEOUT = 10
DESCRIPTION = "Kubernetes container readiness check utility"
USAGE = "Usage: ready.py [-t <timeout>] -c <container_name> " \
        "[-c <container_name> ...]\n" \
        "where\n" \
        "<timeout> - wait for container readiness timeout in min, " \
        "default is " + str(DEF_TIMEOUT) + "\n" \
        "<container_name> - name of the container to wait for\n"


def main(argv):
    """
    Checks if a container is ready or if a job is finished.
    The check is done according to the name of the container, not the name of
    its parent (Job, Deployment, StatefulSet, DaemonSet).

    Args:
        argv: the command line
    """
    # args are a list of container names
    container_names = []
    timeout = DEF_TIMEOUT
    try:
        opts, _args = getopt.getopt(argv, "hc:t:", ["container-name=",
                                                    "timeout=",
                                                    "help"])
        for opt, arg in opts:
            if opt in ("-h", "--help"):
                print("{}\n\n{}".format(DESCRIPTION, USAGE))
                sys.exit()
            elif opt in ("-c", "--container-name"):
                container_names.append(arg)
            elif opt in ("-t", "--timeout"):
                timeout = float(arg)
    except (getopt.GetoptError, ValueError) as exc:
        print("Error parsing input parameters: {}\n".format(exc))
        print(USAGE)
        sys.exit(2)
    if container_names.__len__() == 0:
        print("Missing required input parameter(s)\n")
        print(USAGE)
        sys.exit(2)

    for container_name in container_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = is_ready(container_name)
            if ready is True:
                break
            if time.time() > timeout:
                log.warning("timed out waiting for '%s' to be ready",
                            container_name)
                sys.exit(1)
            else:
                # spread in time potentially parallel execution in multiple
                # containers
                time.sleep(random.randint(5, 11))


if __name__ == "__main__":
    main(sys.argv[1:])
