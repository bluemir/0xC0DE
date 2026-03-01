set -e

TAG=${1:-v3.5.11}
BIN_DIR=$(pwd)/${2:-runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

case $OS in
  darwin) EXT="zip";;
  *)      EXT="tar.gz";;
esac

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
cd $TMP_DIR

curl -LO "https://github.com/etcd-io/etcd/releases/download/$TAG/etcd-$TAG-$OS-$ARCH.$EXT"

case $EXT in
  zip)    unzip "etcd-$TAG-$OS-$ARCH.$EXT";;
  tar.gz) tar -xvf "etcd-$TAG-$OS-$ARCH.$EXT";;
esac

mkdir -p $BIN_DIR
mv etcd-$TAG-$OS-$ARCH/etcd etcd-$TAG-$OS-$ARCH/etcdctl $BIN_DIR
