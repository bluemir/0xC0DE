set -e

TAG=${1:-v0.10.0}
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# Check if URL exists or handle errors? curl -L will fail if 404 with -f hopefully or we just try.
# buildkit release naming: buildkit-v0.10.0.linux-amd64.tar.gz
curl -L https://github.com/moby/buildkit/releases/download/$TAG/buildkit-$TAG.$OS-$ARCH.tar.gz | tar -vxz

mkdir -p $BIN_DIR
mv bin/* $BIN_DIR/

rm -rf $TMP_DIR
