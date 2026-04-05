#!/bin/sh
set -eu

echo "Adding Prometheus Community chart repository"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

OBSERVE_NAMESPACE=$(mage observe)
PROMETHEUS_RELEASE=$(mage prometheus)
PROMETHEUS_VALUES="./config/prometheus/values.yaml"

helm install --namespace $OBSERVE_NAMESPACE $PROMETHEUS_RELEASE oci://ghcr.io/prometheus-community/charts/prometheus -f $PROMETHEUS_VALUES
