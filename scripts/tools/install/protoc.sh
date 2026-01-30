set -e

TAG=${1:-25.1}
BIN_DIR=$(pwd)/${2:-/runtime/tools}

. $(dirname $0)/../detect_os_arch.sh

initArch
initOS

case $OS in
  darwin) PROTOC_OS="osx";;
  *) PROTOC_OS="linux";;
esac

case $ARCH in
  aarch64|arm64) PROTOC_ARCH="aarch_64";;
  *) PROTOC_ARCH="x86_64";;
esac


TMP_DIR=$(mktemp -d)
cd $TMP_DIR

echo "see https://grpc.io/docs/protoc-installation/"

# TODO version arch, os...
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v$TAG/protoc-$TAG-$PROTOC_OS-$PROTOC_ARCH.zip
unzip protoc-$TAG-$PROTOC_OS-$PROTOC_ARCH.zip -d protobuf

mkdir -p $BIN_DIR
rm -rf $BIN_DIR/protobuf
mv protobuf $BIN_DIR/

rm -rf $TMP_DIR
