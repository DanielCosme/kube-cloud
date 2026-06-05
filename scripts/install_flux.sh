#!/bin/sh

set -eu

CLUSTER_NAME=charlie
GITEA_HOST=danicos.dev
echo CLUSTER NAME: $CLUSTER_NAME

flux --kubeconfig ~/.kube/$CLUSTER_NAME \
	bootstrap gitea \
	--token-auth \
	--hostname=$GITEA_HOST \
	--owner=daniel \
	--repository=kube-deploy \
	--private=true \
	--branch=main \
	--personal=true \
	--path=./clusters/$CLUSTER_NAME
