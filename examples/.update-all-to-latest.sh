#!/bin/bash
# .update-all-to-latest.sh is a small helper to update all examples to point to the latest langchaingo release
#
export GOPROXY=direct
export GOWORK=off

syncref="${1:-main}"

for gm in $(find . -name go.mod); do
  (
  cd $(dirname $gm)
  go get -u github.com/tmc/langchaingo@${syncref}
  go mod tidy
)
done
