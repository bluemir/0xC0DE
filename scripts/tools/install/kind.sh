set -e

TAG=${1:-v0.31.0}
BIN_DIR=$(pwd)/${2:-runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

# https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
curl -Lo ./kind "https://kind.sigs.k8s.io/dl/$TAG/kind-$OS-$ARCH"
chmod +x ./kind

mkdir -p $BIN_DIR
mv ./kind $BIN_DIR
