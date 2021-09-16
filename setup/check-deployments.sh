#!/bin/bash

wait_for_resource() {
    local namespace=$1
    local resource_name=$2
    echo ">>> Waiting for ${resource_name}"
    retries=10
    count=0
    while :
    do
        kubectl wait --for=condition=available --timeout=300s -n $namespace $resource_name && break
        sleep 30
        count=$(($count + 1))
        if [[ ${count} -eq ${retries} ]]; then
            echo "No more retries left"
            kubectl get pods -n $namespace
            exit 1
        fi
    done
}

wait_for_deployment() {
    wait_for_resource $1 "deployment/$2"
}

wait_for_deployment "pds-system" "api-server"
wait_for_deployment "pds-system" "api-worker"
