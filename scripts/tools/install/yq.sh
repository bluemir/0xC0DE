set -e

TAG=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS


TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

curl -SsL "https://github.com/mikefarah/yq/releases/download/$TAG/yq_${OS}_${ARCH}" -o yq
chmod +x yq
mv yq $BIN_DIR
