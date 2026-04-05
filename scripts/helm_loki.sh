#!/bin/sh
set -eu

echo "Adding Grafana Community chart repository"
helm repo add grafana-community https://grafana-community.github.io/helm-charts
helm repo update

OBSERVE_NAMESPACE=$(mage observe)
LOKI_VALUES="./config/loki/values.yaml"

# echo Namespace: $OBSERVE_NAMESPACE
# echo Alloy release name: $ALLOY_RELEASE
# echo Alloy values path: $ALLOY_VALUES;

# helm install --namespace $OBSERVE_NAMESPACE loki grafana-community/loki -f $LOKI_VALUES
helm upgrade --namespace $OBSERVE_NAMESPACE loki grafana-community/loki -f $LOKI_VALUES
# helm uninstall loki -n $OBSERVE_NAMESPACE
