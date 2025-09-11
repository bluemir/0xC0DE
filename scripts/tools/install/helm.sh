set -e

TAG=$1
BIN_DIR=$(pwd)/${2:-/runtime/tools}

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

#########################

initArch
initOS

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

cd $TMP_DIR

curl -SsL "https://get.helm.sh/helm-$TAG-$OS-$ARCH.tar.gz" -o helm.tar.gz
tar xf "helm.tar.gz"
mv $OS-$ARCH/helm $BIN_DIR
