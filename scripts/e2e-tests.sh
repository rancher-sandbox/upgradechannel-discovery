#!/bin/bash

$ROOT_DIR/scripts/create-cluster.sh

set -e

kubectl cluster-info --context kind-$CLUSTER_NAME
echo "Sleep to give times to node to populate with all info"
kubectl wait --for=condition=Ready node/$CLUSTER_NAME-control-plane
export EXTERNAL_IP=$(kubectl get nodes -o jsonpath='{.items[].status.addresses[?(@.type == "InternalIP")].address}')
export BRIDGE_IP="172.18.0.1"
kubectl get nodes -o wide
cd $ROOT_DIR/tests &&  ginkgo -r --flake-attempts=3 -v ./e2e
