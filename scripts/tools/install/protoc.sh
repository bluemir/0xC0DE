set -e

BIN_DIR=$(pwd)/${2:-/runtime/tools}


OS=$(go env GOOS)
ARCH=$(go env GOARCH)

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

echo "see https://grpc.io/docs/protoc-installation/"

# TODO version arch, os...
#curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
#unzip protoc-25.1-linux-x86_64.zip -d protobuf
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v31.0/protoc-31.0-osx-aarch_64.zip
unzip protoc-31.0-osx-aarch_64.zip -d protobuf

#unzip protoc-3.15.8-linux-x86_64.zip -d build/tools/protobuf/
#ln -s protobuf/bin/protoc build/tools/protoc

mkdir -p $BIN_DIR
mv protobuf $BIN_DIR/

rm -rf $TMP_DIR
