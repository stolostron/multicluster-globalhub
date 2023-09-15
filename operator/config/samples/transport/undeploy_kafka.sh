#!/bin/bash

KUBECONFIG=${1:-$KUBECONFIG}
currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
rootDir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." ; pwd -P)"
source $rootDir/test/setup/common.sh

# step1: delete transport secret 
targetNamespace=${TARGET_NAMESPACE:-"multicluster-global-hub"}
transportSecret=${TRANSPORT_SECRET_NAME:-"multicluster-global-hub-transport"}
kubectl delete secret ${transportSecret} -n $targetNamespace

# step2: delete kafka topics
kubectl delete -f ${currentDir}/kafka-topics.yaml
waitDisappear "kubectl get kafkatopic spec -n $targetNamespace --ignore-not-found | grep spec || true"
waitDisappear "kubectl get kafkatopic status -n $targetNamespace --ignore-not-found | grep status || true"
waitDisappear "kubectl get kafkatopic event -n $targetNamespace --ignore-not-found | grep status || true"

# step3: delete kafka cluster
kubectl delete -f ${currentDir}/kafka-cluster.yaml
waitDisappear "kubectl -n $targetNamespace get kafka.kafka.strimzi.io/kafka --ignore-not-found"

# step4: delete kafka operator
kubectl delete -f ${currentDir}/kafka-subscription.yaml
kafkaOperator=$(kubectl get deploy -n $targetNamespace | grep strimzi-cluster-operator | awk '{print $1}')
kubectl delete deploy $kafkaOperator -n $targetNamespace
waitDisappear "kubectl get pods -n $targetNamespace | grep strimzi-cluster-operator | grep Running || true"





