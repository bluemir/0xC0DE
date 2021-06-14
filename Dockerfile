ARG VERSION=dev
FROM fedora:34 as build-env

RUN echo "fastestmirror=1" >> /etc/dnf/dnf.conf
RUN dnf install -y \
    make findutils which \
    golang nodejs \
    protobuf protobuf-compiler protobuf-devel \
    && dnf clean all

ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

# pre build
WORKDIR /pre-build

ADD go.mod go.sum package.json yarn.lock Makefile  ./
ADD scripts/makefile.d/ scripts/makefile.d/

## install build tools
RUN make build-tools

## download dependancy
### go
RUN go mod download
### nodejs
RUN yarn install

# build
WORKDIR /src

## for use vendor folder. uncomment next line
#ENV OPTIONAL_BUILD_ARGS="-mod=vendor"

ARG VERSION

## copy source
ADD . /src

RUN make build/0xC0DE

################################################################################
# running image
FROM fedora:34

WORKDIR /
COPY --from=build-env /src/build/0xC0DE /bin/

CMD 0xC0DE

