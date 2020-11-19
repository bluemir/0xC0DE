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

## install build tools
ADD makefile.d/tools.mk makefile.d/tools.mk
RUN make -f makefile.d/tools.mk tools

## download dependancy
ADD go.mod go.sum package.json ./
### go
RUN go mod download
### nodejs
RUN yarn install

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

