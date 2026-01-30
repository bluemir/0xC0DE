set -e

TAG=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

#########################

initArch
initOS

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

curl -SsL "https://dl.k8s.io/release/$TAG/bin/$OS/$ARCH/kubectl" -o kubectl
chmod +x kubectl
mv kubectl $BIN_DIR
