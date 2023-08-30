#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright © 2020 Orange
# Copyright © 2020 Nokia
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
import requests
import socket
from contextlib import closing

from kubernetes import client, config
from kubernetes.client.rest import ApiException

# extract ns from env variable
namespace = os.environ['NAMESPACE']

# setup logging
log = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
handler.setLevel(logging.INFO)
log.addHandler(handler)
log.setLevel(logging.INFO)

config.load_incluster_config()
# use for local testing:
#config.load_kube_config()
coreV1Api = client.CoreV1Api()
api = client.AppsV1Api()
batchV1Api = client.BatchV1Api()

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
        response = api.read_namespaced_daemon_set(
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

def is_pod_ready(pod_name):
    """
    Check if a pod is ready.

    For a pod owned by a Job, it means the Job is complete.
    Otherwise, it means the parent (Deployment, StatefulSet, DaemonSet) is
    running with the right number of replicas

    Args:
        pod_name (str): the name of the pod.

    Returns:
        True if pod is ready, false otherwise
    """
    ready = False
    log.info("Checking if %s is ready", pod_name)
    try:
        response = coreV1Api.list_namespaced_pod(namespace=namespace,
                                                 watch=False)
        for item in response.items:
          if (item.metadata.name.startswith(pod_name)):
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

def is_app_ready(app_name):
    """
    Check if a pod with app-label is ready.

    For a pod owned by a Job, it means the Job is complete.
    Otherwise, it means the parent (Deployment, StatefulSet, DaemonSet) is
    running with the right number of replicas

    Args:
        app_name (str): the app label of the pod.

    Returns:
        True if pod is ready, false otherwise
    """
    ready = False
    log.info("Checking if pod with app-label %s is ready", app_name)
    try:
        response = coreV1Api.list_namespaced_pod(namespace=namespace,
                                                 watch=False)
        for item in response.items:
          if item.metadata.labels.get('app', "NOKEY") == app_name:
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

def service_mesh_job_check(container_name):
    """
    Check if a Job's primary container is complete. Used for ensuring the sidecar can be killed after Job completion.
    Args:
        container_name (str): the name of the Job's primary container.

    Returns:
         True if job's container is in the completed state, false otherwise
    """
    complete = False
    log.info("Checking if %s is complete", container_name)
    try:
        response = coreV1Api.list_namespaced_pod(namespace=namespace, watch=False)
        for item in response.items:
            # container_statuses can be None, which is non-iterable.
            if item.status.container_statuses is None:
                continue
            for container in item.status.container_statuses:
                if container.name == container_name and item.status.phase == "Running":
                    name = read_name(item)
                    log.info("Container Details  %s ", container)
                    log.info("Container Status  %s ", container.state.terminated)

                    if container.state.terminated:
                      log.info("Container Terminated with reason  %s ", container.state.terminated.reason)
                      complete = True

    except ApiException as exc:
        log.error("Exception when calling read_namespaced_job_status: %s\n",
                  exc)
    return complete

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
    api_response = api.read_namespaced_replica_set_status(replicaset,
                                                          namespace)
    deployment_name = read_name(api_response)
    return deployment_name

def check_socket(host, port):
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as sock:
        if sock.connect_ex((host, port)) == 0:
            print("Port is open")
            return True
        else:
            print("Port is not open")
            return False

def quitquitquit_post(apiurl):
    URL = apiurl
    if check_socket("127.0.0.1", 15020) is False:
        log.info("no sidecar exists, exiting")
        return True
    response = requests.post(url = URL)
    responseStatus = response.ok
    try:
        if responseStatus is True:
            log.info("quitquitquit returned True")
            return True
        else:
            log.info("quitquitquit returned False")
            return False
    except:
        log.info("quitquitquit call failed with exception")

DEF_TIMEOUT = 10
DEF_URL = "http://127.0.0.1:15020/quitquitquit"
DESCRIPTION = "Kubernetes container readiness check utility"
USAGE = "Usage: ready.py [-t <timeout>] -c <container_name> .. | -j <job_name> .. " \
        "| -p <pod_name> .. | -a <app_name> .. \n" \
        "where\n" \
        "<timeout> - wait for container readiness timeout in min, " \
        "default is " + str(DEF_TIMEOUT) + "\n" \
        "<container_name> - name of the container to wait for\n" \
        "<pod_name> - name of the pod to wait for\n" \
        "<app_name> - app label of the pod to wait for\n" \
        "<job_name> - name of the job to wait for\n"


def main(argv):
    """
    Checks if a container or pod is ready, 
    if a job is finished or if the main container of a job has completed.
    The check is done according to the name of the container op pod,
    not the name of its parent (Job, Deployment, StatefulSet, DaemonSet).

    Args:
        argv: the command line
    """
    # args are a list of container names
    container_names = []
    pod_names = []
    app_names = []
    job_names = []
    service_mesh_job_container_names = []
    timeout = DEF_TIMEOUT
    url = DEF_URL
    try:
        opts, _args = getopt.getopt(argv, "hj:c:p:a:t:s:u:", ["container-name=",
                                                    "pod-name=",
                                                    "app-name=",
                                                    "timeout=",
                                                    "service-mesh-check=",
                                                    "url=",
                                                    "job-name=",
                                                    "help"])
        for opt, arg in opts:
            if opt in ("-h", "--help"):
                print("{}\n\n{}".format(DESCRIPTION, USAGE))
                sys.exit()
            elif opt in ("-c", "--container-name"):
                container_names.append(arg)
            elif opt in ("-p", "--pod-name"):
                pod_names.append(arg)
            elif opt in ("-a", "--app-name"):
                app_names.append(arg)
            elif opt in ("-j", "--job-name"):
                job_names.append(arg)
            elif opt in ("-s", "--service-mesh-check"):
                service_mesh_job_container_names.append(arg)
            elif opt in ("-u", "--url"):
                url = arg
            elif opt in ("-t", "--timeout"):
                timeout = float(arg)
    except (getopt.GetoptError, ValueError) as exc:
        print("Error parsing input parameters: {}\n".format(exc))
        print(USAGE)
        sys.exit(2)
    if container_names.__len__() == 0 and job_names.__len__() == 0 and pod_names.__len__() == 0 \
       and app_names.__len__() == 0 and service_mesh_job_container_names.__len__() == 0:
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
    for pod_name in pod_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = is_pod_ready(pod_name)
            if ready is True:
                break
            if time.time() > timeout:
                log.warning("timed out waiting for '%s' to be ready",
                            pod_name)
                sys.exit(1)
            else:
                # spread in time potentially parallel execution in multiple
                # containers
                time.sleep(random.randint(5, 11))
    for app_name in app_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = is_app_ready(app_name)
            if ready is True:
                break
            if time.time() > timeout:
                log.warning("timed out waiting for '%s' to be ready",
                            pod_name)
                sys.exit(1)
            else:
                # spread in time potentially parallel execution in multiple
                # containers
                time.sleep(random.randint(5, 11))
    for job_name in job_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = is_job_complete(job_name)
            if ready is True:
                break
            if time.time() > timeout:
                log.warning("timed out waiting for '%s' to be ready",
                            job_name)
                sys.exit(1)
            else:
                # spread in time potentially parallel execution in multiple
                # containers
                time.sleep(random.randint(5, 11))
    for service_mesh_job_container_name in service_mesh_job_container_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = service_mesh_job_check(service_mesh_job_container_name)
            if ready is True:
                sideCarKilled = quitquitquit_post(url)
                if sideCarKilled is True:
                    log.info("Side Car Killed through QuitQuitQuit API")
                else:
                    log.info("Side Car Failed to be Killed through QuitQuitQuit API")
                break
            if time.time() > timeout:
                log.warning("timed out waiting for '%s' to be ready",
                            service_mesh_job_container_name)
                sys.exit(1)
            else:
                # spread in time potentially parallel execution in multiple
                # containers
                time.sleep(random.randint(5, 11))

if __name__ == "__main__":
    main(sys.argv[1:])
