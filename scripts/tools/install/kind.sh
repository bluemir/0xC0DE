set -e

BIN_DIR=$(pwd)/${2:-/runtime/tools}

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries

[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.24.0/kind-linux-amd64
# For ARM64
[ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.24.0/kind-linux-arm64
chmod +x ./kind

mkdir -p $BINDIR
sudo mv ./kind $(BIN_DIR)

rm -rf $TMP_DIR
