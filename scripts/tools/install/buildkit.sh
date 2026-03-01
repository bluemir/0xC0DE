set -e

TAG=${1:-v0.27.1}
BIN_DIR=$(pwd)/${2:-runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

# buildkit release naming: buildkit-v0.27.1.linux-amd64.tar.gz
curl -L https://github.com/moby/buildkit/releases/download/$TAG/buildkit-$TAG.$OS-$ARCH.tar.gz | tar -vxz

mkdir -p $BIN_DIR
mv bin/* $BIN_DIR/
