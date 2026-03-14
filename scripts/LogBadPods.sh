#!/bin/bash
#
# Log the details for each pod in a non-Running/Completed state.
#
set -e

# Verify connectivity to the cluster.
kubectl cluster-info >/dev/null

TS="$(date +%Y%m%dT%H%M%S)"
LOGFILE="bad_pods_$TS.txt"
echo "Saving pod information to: $LOGFILE"

mapfile -t BAD_STATE < <(kubectl get --no-headers po -A | grep -vE 'Running|Completed' | awk '{ print $1 "/" $2 }')

exec > $LOGFILE 2>&1
for pod in "${BAD_STATE[@]}"; do
    namespace="${pod%%/*}"
    name="${pod##*/}"

    echo "----- $(date): $namespace/$name Pod Details -----"
    kubectl -n $namespace describe pod $name
    echo
done
