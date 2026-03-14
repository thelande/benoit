#!/bin/bash
#
# Check the ready state of the kubernetes nodes. List out nodes not in Ready
# state and exit 1 if any NotReady nodes are found.
#
set -e

# Verify connectivity to the cluster.
kubectl cluster-info >/dev/null

mapfile -t NOT_READY_NODES < <(kubectl get --no-headers nodes | grep NotReady | awk '{ print $1 }')
if [[ "${#NOT_READY_NODES[@]}" -gt 0 ]]; then
    echo "Kubernetes nodes not in Ready state:"
    echo "${NOT_READY_NODES[@]}"
    exit 1
fi

exit 0
