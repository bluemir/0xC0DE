set -e

TAG=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS


TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

cd $TMP_DIR

curl -SsL "https://get.helm.sh/helm-$TAG-$OS-$ARCH.tar.gz" -o helm.tar.gz
tar xf "helm.tar.gz"
mv $OS-$ARCH/helm $BIN_DIR
