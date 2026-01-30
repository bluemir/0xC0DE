set -e

BIN_DIR=$(pwd)/${2:-/runtime/tools}
PAGE_VERSION=v0.24.0

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
curl -Lo ./kind "https://kind.sigs.k8s.io/dl/$PAGE_VERSION/kind-$OS-$ARCH"
chmod +x ./kind

mkdir -p $BIN_DIR
mv ./kind $BIN_DIR

rm -rf $TMP_DIR
