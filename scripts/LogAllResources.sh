#!/bin/bash
#
# Log information about all Kubernetes resources.
#
set -e

# Verify connectivity to the cluster.
kubectl cluster-info >/dev/null

TS="$(date +%Y%m%dT%H%M%S)"
LOGFILE="k8s_resources_$TS.txt"
echo "Saving resource information to: $LOGFILE"

exec > $LOGFILE 2>&1
TYPES=(nodes pods job deployments daemonset replicaset statefulset service pvc pv configmap secret)

for t in "${TYPES[@]}"; do
    echo "----- $(date): $t -----"
    kubectl get $t -A
    echo
done
