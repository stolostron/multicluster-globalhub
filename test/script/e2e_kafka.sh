#!/bin/bash

CURRENT_DIR=$(
  cd "$(dirname "$0")" || exit
  pwd
)
# shellcheck source=/dev/null
source "$CURRENT_DIR/util.sh"

KAFKA_KUBECONFIG=${1:-$KUBECONFIG}  # install the kafka
SECRET_KUBECONFIG=${2:-$KUBECONFIG} # generate the crenditial secret

echo "KAFKA_KUBECONFIG=$KAFKA_KUBECONFIG"
echo "SECRET_KUBECONFIG=$SECRET_KUBECONFIG"

start_time=$(date +%s)
echo -e "\r${BOLD_GREEN}[ START - $(date +"%T") ] Install Kafka $NC"

# check the transport secret
secret_name=${TRANSPORT_SECRET_NAME:-"multicluster-global-hub-transport"}
secret_namespace=${TRANSPORT_SECRET_NAMESPACE:-"multicluster-global-hub"}
kafka_namespace=${KAFKA_NAMESPACE:-"kafka"}

kubectl create ns "$secret_namespace" --dry-run=client -oyaml | kubectl --kubeconfig "$SECRET_KUBECONFIG" apply -f -
if kubectl get secret "$secret_name" -n "$secret_namespace" --kubeconfig "$SECRET_KUBECONFIG"; then
  echo "secret_name $secret_name already exists in $secret_namespace namespace"
  exit 0
fi

# create all the resource in cluster KUBECONFIG
kubectl create namespace "$kafka_namespace" --kubeconfig "$KAFKA_KUBECONFIG" --dry-run=client -o yaml | kubectl apply -f - --kubeconfig "$KAFKA_KUBECONFIG"

# deploy kafka operator
kubectl -n "$kafka_namespace" create -f "https://strimzi.io/install/latest?namespace=$kafka_namespace" --kubeconfig "$KAFKA_KUBECONFIG"
retry "(kubectl get pods -n $kafka_namespace --kubeconfig $KAFKA_KUBECONFIG -l name=strimzi-cluster-operator | grep Running)" 60

echo "Kafka operator is ready"

node_port_host=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' --kubeconfig "$KAFKA_KUBECONFIG" | sed -e 's#^https\?://##' -e 's/:.*//')
sed -i -e "s;NODE_PORT_HOST;$node_port_host;" "$TEST_DIR"/manifest/kafka/kafka-cluster/kafka-cluster.yaml
# deploy kafka cluster
kubectl apply -k "$TEST_DIR"/manifest/kafka/kafka-cluster -n "$kafka_namespace" --kubeconfig "$KAFKA_KUBECONFIG"

wait_cmd "kubectl get kafka kafka -n $kafka_namespace --kubeconfig $KAFKA_KUBECONFIG -o jsonpath='{.status.listeners[1]}' | grep bootstrapServers"

echo "Kafka cluster is ready"

# byo kafkatopic and kafkauser
wait_cmd "kubectl get kafkatopic gh-spec -n $kafka_namespace --kubeconfig $KAFKA_KUBECONFIG | grep -C 1 True"
wait_cmd "kubectl get kafkatopic gh-status -n $kafka_namespace --kubeconfig $KAFKA_KUBECONFIG | grep -C 1 True"
byo_user=global-hub-byo-user
wait_cmd "kubectl get kafkauser $byo_user -n $kafka_namespace --kubeconfig $KAFKA_KUBECONFIG | grep -C 1 True"
echo "Kafka topic and user is ready"

# generate transport secret for standalone agent
export KAFKA_NAMESPACE=$kafka_namespace
bash "$TEST_DIR/manifest/standalone-agent/generate_transport_config.sh" "$KAFKA_KUBECONFIG" "$SECRET_KUBECONFIG"
echo "Kafka standalone secret is ready! KUBECONFIG=$SECRET_KUBECONFIG"

echo -e "\r${BOLD_GREEN}[ END - $(date +"%T") ] Install Kafka ${NC} $(($(date +%s) - start_time)) seconds"
