set -e

VERSION=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

OS=$(go env GOOS)
ARCH=$(go env GOARCH)

TMP_DIR=$(mktemp -d)
cd $TMP_DIR


curl -LO "https://github.com/etcd-io/etcd/releases/download/$VERSION/etcd-$VERSION-$OS-$ARCH.tar.gz"
tar -xvf "etcd-$VERSION-$OS-$ARCH.tar.gz"


mkdir -p $BIN_DIR
mv etcd-$VERSION-$OS-$ARCH/etcd etcd-$VERSION-$OS-$ARCH/etcdctl $BIN_DIR

rm -rf $TMP_DIR
