set -e

BIN_DIR=$(pwd)/${2:-/runtime/tools}

OS=$(go env GOOS)
ARCH=$(go env GOARCH)

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

curl -L https://github.com/moby/buildkit/releases/download/v0.10.0/buildkit-v0.10.0.$OS-$ARCH.tar.gz | tar -vxz

mkdir -p $BIN_DIR
mv bin/* $BIN_DIR

rm -rf $TMP_DIR
