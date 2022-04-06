set -e

export GOBIN=$PWD/bin/

TMP_DIR=$(mktemp -d)
cd $TMP_DIR

go mod init temp

go get $1
go install $1 #go 1.18

rm -rf $TMP_DIR

echo "$1 installed"
