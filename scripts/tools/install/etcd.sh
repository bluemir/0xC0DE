set -e

TAG=${1:-v3.5.11}
BIN_DIR=$(pwd)/${2:-runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

curl -LO "https://github.com/etcd-io/etcd/releases/download/$TAG/etcd-$TAG-$OS-$ARCH.tar.gz"
tar -xvf "etcd-$TAG-$OS-$ARCH.tar.gz"

mkdir -p $BIN_DIR
mv etcd-$TAG-$OS-$ARCH/etcd etcd-$TAG-$OS-$ARCH/etcdctl $BIN_DIR
