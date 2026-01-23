
#!/bin/bash
# .build-all-examples.sh is a small helper to build all examples to point to the latest langchaingo release
#
export GOPROXY=direct
export GOWORK=off

mkdir -p ../.build

for gm in $(find . -name go.mod); do
  (
    echo "Building $(dirname $gm)"
    cd $(dirname $gm)
    go build -o ../../.build/$(basename $(dirname $gm)) ./...
) &
done
wait
