#!/bin/sh
set -eu

echo "Adding Grafana Helm chart repository"
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

OBSERVE_NAMESPACE=$(mage observe)
ALLOY_RELEASE=$(mage alloy)
ALLOY_VALUES="./config/alloy/values.yaml"

echo Namespace: $OBSERVE_NAMESPACE
echo Alloy release name: $ALLOY_RELEASE
echo Alloy values path: $ALLOY_VALUES;
# helm install --namespace $OBSERVE_NAMESPACE $ALLOY_RELEASE grafana/alloy -f $ALLOY_VALUES
# helm uninstall $ALLOY_RELEASE -n $OBSERVE_NAMESPACE
helm upgrade --namespace $OBSERVE_NAMESPACE $ALLOY_RELEASE grafana/alloy -f $ALLOY_VALUES
