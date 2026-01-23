#!/bin/bash
# .update-all-to-latest.sh is a small helper to update all examples to point to the latest langchaingo release
#
export GOPROXY=direct
export GOWORK=off

syncref="${1:-latest}"

for gm in $(find . -name go.mod); do
  (
  echo "Tidying $(dirname $gm)"
  cd $(dirname $gm)
  go mod tidy
) &
done
wait
