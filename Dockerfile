ARG VERSION=dev
FROM fedora:40 as build-env

RUN echo "fastestmirror=1" >> /etc/dnf/dnf.conf
RUN dnf install -y \
    make findutils which \
    golang nodejs \
    protobuf protobuf-compiler protobuf-devel \
    && dnf clean all

ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

# pre build
WORKDIR /src

COPY Makefile ./
COPY scripts/ scripts/

## install build tools
RUN make build-tools 2>/dev/null

## download dependancy

COPY go.mod go.sum package.json package-lock.json ./

### go
RUN go mod download
### nodejs
RUN npm install

# build
# WORKDIR /src

## for use vendor folder. uncomment next line
#ENV OPTIONAL_BUILD_ARGS="-mod=vendor"
ENV OPTIONAL_WEB_BUILD_ARGS="--minify"
ENV OPTIONAL_BUILD_ARGS="--tags embed"

ARG VERSION

## copy source
COPY . /src

RUN make build/0xC0DE

################################################################################
# running image
FROM fedora:40

WORKDIR /
COPY --from=build-env /src/build/0xC0DE /bin/

CMD 0xC0DE

