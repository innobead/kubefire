#!/usr/bin/env bash

function enter() {
    read -n 1 key
    clear

    if [[ "$key" != "" ]]; then
        exit 1
    fi

    if [[ $@ != echo* ]]; then
        echo $@

    fi
    
    "$@"
}

enter echo "1. install the first K3s cluster and record it's kubeconfig path w/ unique pod and service IPs"
enter kubefire cluster create k3s-1 --bootstrapper=k3s --extra-options=server_install_options="--cluster-cidr='10.42.0.0/16',--service-cidr='10.43.0.0/16'" --force
enter export k3s1_path=$(kubefire cluster env k3s-1 --path-only)

enter echo "2. install the second K3s cluster and record it's kubeconfig path w/ unique pod and service IPs"
enter kubefire cluster create k3s-2 --bootstrapper=k3s --extra-options=server_install_options="--cluster-cidr='10.52.0.0/16',--service-cidr='10.53.0.0/16'" --force
enter export k3s2_path=$(kubefire cluster env k3s-2 --path-only)

enter echo "3. install submariner broker - CRDs and API resources"
enter subctl deploy-broker --kubeconfig "$k3s1_path"

enter echo "4. for k3s-1 cluster, label the node as gateway engine, join the submariner broker"
enter export KUBECONFIG=$k3s1_path
enter kubectl label node $(kubectl get nodes -o=jsonpath="{.items[0].metadata.labels.kubernetes\.io\/hostname}") submariner.io/gateway=true
enter subctl join broker-info.subm --clusterid k3s-1 --clustercidr=10.42.0.0/16 --servicecidr=10.43.0.0/16 --disable-nat

enter echo "4. for k3s-2 cluster, label the node as gateway engine, join the submariner broker"
enter export KUBECONFIG=$k3s2_path
enter kubectl label node $(kubectl get nodes -o=jsonpath="{.items[0].metadata.labels.kubernetes\.io\/hostname}") submariner.io/gateway=true
enter subctl join broker-info.subm --clusterid k3s-2 --clustercidr=10.52.0.0/16 --servicecidr=10.53.0.0/16 --disable-nat

enter echo "5. Wait until submariner endpoints connected"
enter export KUBECONFIG=$k3s1_path
enter watch subctl show all

enter echo "6. run a nginx pod in the first cluster"
enter export KUBECONFIG=$k3s1_path
enter kubectl delete pod nginx
enter kubectl run nginx --image=nginx
enter watch kubectl get pod nginx -o wide
enter export nginx_pod_ip=$(kubectl get pod nginx -o=jsonpath="{.status.podIPs[0].ip}")

enter echo "7. run a busybox pod in the second cluster, then ping the nginx pod"
enter export KUBECONFIG=$k3s2_path
enter kubectl delete pod busybox
enter kubectl run busyboxi -it --image=busybox --restart=Never -- ping $nginx_pod_ip



