#!/bin/bash

set -euo pipefail

CURRENT_DIR=$(cd "$(dirname "$0")" || exit;pwd)
# shellcheck source=/dev/null
source "$CURRENT_DIR/common.sh"

export CURRENT_DIR
export GH_NAME="global-hub"
export GH_CTX="kind-global-hub"
export MH_NUM=${MH_NUM:-2}
export MC_NUM=${MC_NUM:-1}

# setup kubeconfig
export KUBE_DIR=${CURRENT_DIR}/kubeconfig
check_dir "$KUBE_DIR"
export KUBECONFIG=${KUBECONFIG:-${KUBE_DIR}/kind-clusters}

# Init clusters
echo -e "$BLUE creating clusters $NC"
start_time=$(date +%s)
kind_cluster "$GH_NAME" 2>&1 &
for i in $(seq 1 "${MH_NUM}"); do
  hub_name="hub$i"
  kind_cluster "$hub_name" 2>&1 &
  for j in $(seq 1 "${MC_NUM}"); do
    cluster_name="hub$i-cluster$j"
    kind_cluster "$cluster_name" 2>&1 &
  done
done
wait
end_time=$(date +%s)
sum_time=$((end_time - start_time))
echo -e "$YELLOW creating clusters:$NC $sum_time seconds"

# Init hub resources
start_time=$(date +%s)

# GH 
export GH_KUBECONFIG=$KUBE_DIR/kind-$GH_NAME
start_time=$(date +%s)
node_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${GH_NAME}-control-plane)
kubectl --kubeconfig "$GH_KUBECONFIG" config set-cluster kind-$GH_NAME --server="https://$node_ip:6443"

# init resources
echo -e "$BLUE initilize resources $NC"
install_crds $GH_CTX 2>&1 &    # router, mch(not needed for the managed clusters)
enable_service_ca $GH_CTX $GH_NAME "$CURRENT_DIR/hoh" 2>&1 &

for i in $(seq 1 "${MH_NUM}"); do
  install_crds "kind-hub$i" 2>&1 &
  for j in $(seq 1 "${MC_NUM}"); do
    install_crds "kind-hub$i-cluster$j" 2>&1 &
  done
done

# # install middlewares in async mode
bash "$CURRENT_DIR"/hoh/postgres/postgres_setup.sh "$GH_KUBECONFIG" 2>&1 &
bash "$CURRENT_DIR"/hoh/kafka/kafka_setup.sh "$GH_KUBECONFIG" 2>&1 &

# init hubs
echo -e "$BLUE initializing hubs $NC"
init_hub $GH_CTX &
for i in $(seq 1 "${MH_NUM}"); do
  init_hub "kind-hub$i" &
done

wait 
end_time=$(date +%s)
echo -e "$YELLOW Initializing hubs:$NC $((end_time - start_time)) seconds"

# Importing clusters
start_time=$(date +%s)

# join spoken clusters, if the cluster has been joined, then skip
echo -e "$BLUE importing clusters $NC"
for i in $(seq 1 "${MH_NUM}"); do
  mh_ctx="kind-hub$i"
  join_cluster $GH_CTX $mh_ctx 2>&1 &  # join to global hub
  for j in $(seq 1 "${MC_NUM}"); do
    mc_ctx="kind-hub$i-cluster$j"
    join_cluster $mh_ctx $mc_ctx 2>&1 & # join to managed hub
  done
done
wait 

end_time=$(date +%s)
echo -e "$YELLOW Importing cluster:$NC $((end_time - start_time)) seconds"

# Install app and policy
start_time=$(date +%s)

# app
echo -e "$BLUE deploying app $NC"
for i in $(seq 1 "${MH_NUM}"); do
  init_app $GH_CTX "kind-hub$i" 2>&1 &
  for j in $(seq 1 "${MC_NUM}"); do
    init_app "kind-hub$i" "kind-hub$i-cluster$j" 2>&1 &
  done
done

# policy
echo -e "$BLUE deploying policy $NC"
for i in $(seq 1 "${MH_NUM}"); do
  init_policy $GH_CTX "kind-hub$i" 2>&1 &
  for j in $(seq 1 "${MC_NUM}"); do
    init_policy "kind-hub$i" "kind-hub$i-cluster$j" 2>&1 &
  done
done

wait 
end_time=$(date +%s)
echo -e "$YELLOW App and policy:$NC $((end_time - start_time)) seconds"

# wait middleware to be ready
wait_appear "kubectl get kafka kafka -n multicluster-global-hub -o jsonpath='{.status.listeners[1].certificates[0]}' --context $GH_CTX" 1200
wait_appear "kubectl get secret hoh-pguser-postgres -n hoh-postgres --ignore-not-found=true --context $GH_CTX" 1200

echo -e "$BLUE enable the clusters $NC"
for i in $(seq 1 "${MH_NUM}"); do
  enable_cluster $GH_CTX "kind-hub$i" 2>&1 &
  for j in $(seq 1 "${MC_NUM}"); do
    enable_cluster "kind-hub$i" "kind-hub$i-cluster$j" 2>&1 &
  done
done

wait

#need the following labels to enable deploying agent in leaf hub cluster
for i in $(seq 1 "${MH_NUM}"); do
  echo -e "$GREEN [Access the Clusters]: export KUBECONFIG=$KUBE_DIR/kind-hub$i $NC"
  for j in $(seq 1 "${MC_NUM}"); do
    echo -e "$GREEN [Access the Clusters]: export KUBECONFIG=$KUBE_DIR/kind-hub$i-cluster$j $NC"
  done
done
echo -e "$GREEN [Access the Clusters]: export KUBECONFIG=$KUBECONFIG $NC"
