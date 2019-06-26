ARG VERSION=dev
ARG BIN_NAME
FROM fedora:29 as build-env

RUN  yum install -y \
    golang make findutils \
    && yum clean all

ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

# build

## tools
### rice
RUN  go get -u github.com/GeertJohan/go.rice/rice

## dep
ADD go.mod go.sum ./
RUN go mod download

ARG VERSION

## copy source
ADD . /src
WORKDIR /src

RUN make build/$BIN_NAME

################################################################################
# running image
FROM fedora:29

WORKDIR /
COPY --from=build-env /src/build/$BIN_NAME /bin/$BIN_NAME

ENTRYPOINT ["$BIN_NAME"]

