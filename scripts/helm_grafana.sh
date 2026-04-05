#!/bin/sh
set -eu

echo "Adding Grafana Community chart repository"
helm repo add grafana-community https://grafana-community.github.io/helm-charts
helm repo update

OBSERVE_NAMESPACE=$(mage observe)
GRAFANA_RELEASE=$(mage grafana)
GRAFANA_VALUES="./config/grafana/values.yaml"

helm install --namespace $OBSERVE_NAMESPACE $GRAFANA_RELEASE grafana-community/grafana -f $GRAFANA_VALUES
# helm upgrade --namespace $OBSERVE_NAMESPACE $GRAFANA_RELEASE grafana-community/grafana -f $GRAFANA_VALUES
# helm uninstall $GRAFANA_RELEASE -n $OBSERVE_NAMESPACE

# Get admin password
# kubectl get secret --namespace observe grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
