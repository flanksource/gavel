#!/usr/bin/env bash
# Copies agent-substrate's generated credential provider API next to this script, so gavel
# credential-provider serves Substrate without depending on its Go modules or their k8s version.
set -o errexit -o nounset -o pipefail

SUBSTRATE=go.ate.dev/substrate@v0.5.1-0.20261016164544-3a2ca6235748

dest="$(cd "$(dirname "$0")" && pwd)"
moddir() { GOWORK=off GOFLAGS=-mod=mod go mod download -json "$1" | sed -n 's/^\t"Dir": "\(.*\)",$/\1/p'; }

substrate="$(moddir "${SUBSTRATE}")"
mkdir -p "${dest}/credproviderpb"
# The credential provider API substrate's egress gateway calls, which gavel credential-provider serves.
for file in credprovider.proto credprovider.pb.go credprovider_grpc.pb.go; do
  install -m 0644 "${substrate}/pkg/proto/credproviderpb/${file}" "${dest}/credproviderpb/${file}"
done
printf 'Copied by sync.sh from %s (credproviderpb).\n' "${SUBSTRATE}" > "${dest}/SOURCE"
