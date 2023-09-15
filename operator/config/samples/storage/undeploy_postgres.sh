#!/bin/bash

currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
rootDir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." ; pwd -P)"
source $rootDir/test/setup/common.sh

# step1: delete transport secret
targetNamespace=${TARGET_NAMESPACE:-"multicluster-global-hub"}
storageSecret=${STORAGE_SECRET_NAME:-"multicluster-global-hub-storage"}
kubectl delete secret $storageSecret -n $targetNamespace
echo "deletes storage secret $storageSecret from namespace $targetNamespace"

# step2: delete postgres cluster
kubectl delete -f ${currentDir}/postgres-cluster.yaml

superuserSecret="postgres-pguser-postgres"
waitDisappear "kubectl get secret $superuserSecret -n $targetNamespace --ignore-not-found=true"
echo "postgres cluster is deleted"

# step3: delete postgres operator
kubectl delete -f ${currentDir}/postgres-subscription.yaml
# kubectl delete subscription.operators.coreos.com crunchy-postgres-operator -n $targetNamespace
csv=$(kubectl get clusterserviceversion -n $targetNamespace | grep postgresoperator | awk '{print $1}')
kubectl delete clusterserviceversion $csv -n $targetNamespace
waitDisappear "kubectl get deploy pgo -n $targetNamespace --ignore-not-found=true"
echo "postgres operator: pgo is deleted"