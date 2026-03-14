#!/bin/bash
#
set -e

# Verify connectivity to the cluster.
kubectl cluster-info >/dev/null

exec >&2

mapfile -t BAD_STATE < <(kubectl get --no-headers po -A | grep -vE 'Running|Completed' | awk '{ print $1 "/" $2 "/" $4 }')
if [[ "${#BAD_STATE[@]}" -gt 0 ]]; then
    echo "Pods not in Running state:"
    for pod in "${BAD_STATE[@]}"; do
        echo "  $pod"
    done
    exit 1
fi

exit 0
