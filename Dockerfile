ARG VERSION=dev
FROM fedora:33 as build-env

RUN echo "fastestmirror=1" >> /etc/dnf/dnf.conf
RUN dnf install -y \
    make findutils which \
    golang nodejs \
    && dnf clean all

ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

# build
WORKDIR /src

ADD go.mod go.sum package.json yarn.lock ./
ADD Makefile.d/tools.mk Makefile.d/tools.mk

## download dependancy
### go
RUN go mod download
### nodejs
RUN yarn install

## install build tools
RUN make -f Makefile.d/tools.mk tools

## for use vendor folder. uncomment next line
#ENV OPTIONAL_BUILD_ARGS="-mod=vendor"

ARG VERSION

## copy source
ADD . /src

ARG APP_NAME
RUN make build/$APP_NAME

################################################################################
# running image
FROM fedora:33

WORKDIR /
ARG APP_NAME
ENV APP_NAME $APP_NAME
COPY --from=build-env /src/build/$APP_NAME /bin/$APP_NAME

CMD $APP_NAME

