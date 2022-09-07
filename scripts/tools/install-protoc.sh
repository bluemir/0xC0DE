set -e

BINDIR=$(pwd)/build/tools

OS=$(go env GOOS)
ARCH=$(go env GOARCH)

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

echo "see https://grpc.io/docs/protoc-installation/"

# TODO version arch, os...
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
unzip protoc-3.15.8-linux-x86_64.zip -d protobuf

#unzip protoc-3.15.8-linux-x86_64.zip -d build/tools/protobuf/
#ln -s protobuf/bin/protoc build/tools/protoc

mkdir -p $BINDIR
mv protobuf $BINDIR/

rm -rf $TMP_DIR
