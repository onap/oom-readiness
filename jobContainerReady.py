#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright Â© 2020 Orange
# Copyright Â© 2020 Nokia
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
api = client.AppsV1Api(client.ApiClient(configuration))
batchV1Api = client.BatchV1Api(client.ApiClient(configuration))


def is_container_complete(container_name):
    """
    Check if Job is complete.
    Args:
        container_name (str): the name of the Job.

    Returns:
         True if job is complete, false otherwise
    """
    complete = False
    log.info("Checking if %s is complete", container_name)
    try:
        response2 = coreV1Api.list_namespaced_pod(namespace=namespace, watch=False)
        for item in response2.items:
            # container_statuses can be None, which is non-iterable.
            if item.status.container_statuses is None:
                continue
                log.info("Triggered Continue block")
            for container in item.status.container_statuses:
                if container.name == container_name:
                    name = read_name(item)
                    log.info("container detail  %s ", container)
                    log.info("container detail  %s ", container.state.terminated)
                    if container.state.terminated is None:
                        continue
                    log.info("container detail  %s ", container.state.terminated.reason)
                    if container.state.terminated.reason == 'Completed':
                        complete = True
                        log.info("%s is complete", container_name)
                    else:
                        log.info("%s is NOT complete", container_name)
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

def quitquitquit_post():
    URL = "http://127.0.0.1:15020/quitquitquit"
    r = requests.post(url = URL)
    responseStatus = r.ok
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
DESCRIPTION = "Kubernetes container readiness check utility"
USAGE = "Usage: jobContainerReady.py [-t <timeout>] -c <container_name> .. | -j <job_name> .. \n" \
        "where\n" \
        "<timeout> - wait for container readiness timeout in min, " \
        "default is " + str(DEF_TIMEOUT) + "\n" \
        "<container_name> - name of the container to wait for\n" \
        "<job_name> - name of the job to wait for\n"


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
    job_names = []
    timeout = DEF_TIMEOUT
    try:
        opts, _args = getopt.getopt(argv, "hj:c:t:", ["container-name=",
                                                    "timeout=",
                                                    "job-name=",
                                                    "help"])
        for opt, arg in opts:
            if opt in ("-h", "--help"):
                print("{}\n\n{}".format(DESCRIPTION, USAGE))
                sys.exit()
            elif opt in ("-c", "--container-name"):
                container_names.append(arg)
            elif opt in ("-j", "--job-name"):
                job_names.append(arg)
            elif opt in ("-t", "--timeout"):
                timeout = float(arg)
    except (getopt.GetoptError, ValueError) as exc:
        print("Error parsing input parameters: {}\n".format(exc))
        print(USAGE)
        sys.exit(2)
    if container_names.__len__() == 0 and job_names.__len__() == 0:
        print("Missing required input parameter(s)\n")
        print(USAGE)
        sys.exit(2)

    for container_name in container_names:
        timeout = time.time() + timeout * 60
        while True:
            ready = is_container_complete(container_name)
            if ready is True:
                sidecarKilled = quitquitquit_post()
                log.info("Side Car Killed through QuitQuitQuit API")
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

