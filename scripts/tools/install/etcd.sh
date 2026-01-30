set -e

VERSION=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

TMP_DIR=$(mktemp -d)
cd $TMP_DIR


curl -LO "https://github.com/etcd-io/etcd/releases/download/$VERSION/etcd-$VERSION-$OS-$ARCH.tar.gz"
tar -xvf "etcd-$VERSION-$OS-$ARCH.tar.gz"


mkdir -p $BIN_DIR
mv etcd-$VERSION-$OS-$ARCH/etcd etcd-$VERSION-$OS-$ARCH/etcdctl $BIN_DIR

rm -rf $TMP_DIR
