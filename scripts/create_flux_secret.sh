#!/bin/sh

set -eu

if [ -z "${AGE_KEY_NO_PQ}" ]; then
	echo "unbound variable"
fi
if [ ! -f "${AGE_KEY_NO_PQ}" ]; then
	echo "Error: ${AGE_KEY_NO_PQ} file does not exist"
	exit 1
fi

cat $AGE_KEY_NO_PQ | kubectl --kubeconfig ~/.kube/charlie create secret generic sops-age --namespace=flux-system --from-file=age.agekey=/dev/stdin
