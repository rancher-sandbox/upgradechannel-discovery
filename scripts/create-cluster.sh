#!/bin/bash

KUBE_VERSION=${KUBE_VERSION:-v1.22.7}
CLUSTER_NAME="${CLUSTER_NAME:-upgradechannel-discovery}"

if ! kind get clusters | grep "$CLUSTER_NAME"; then
cat << EOF > kind.config
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: kindest/node:$KUBE_VERSION
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
EOF
    kind create cluster --name $CLUSTER_NAME --config kind.config
    rm -rf kind.config
fi