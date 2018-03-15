# ============================================================
# Dockerfile used to build the nrod-cif microservice
# ============================================================

ARG arch=amd64
ARG goos=linux

# ============================================================
# Build container containing our pre-pulled libraries.
# As this changes rarely it means we can use the cache between
# building each microservice.
FROM golang:alpine as build

# The golang alpine image is missing git so ensure we have additional tools
RUN apk add --no-cache \
      curl \
      git

# We want to build our final image under /dest
# A copy of /etc/ssl is required if we want to use https datasources
RUN mkdir -p /dest/etc &&\
    cp -rp /etc/ssl /dest/etc/

# Ensure we have the libraries - docker will cache these between builds
RUN go get -v \
      flag \
      github.com/coreos/bbolt/... \
      github.com/gorilla/mux \
      github.com/peter-mount/golib/codec \
      github.com/peter-mount/golib/rest \
      github.com/peter-mount/golib/statistics \
      github.com/peter-mount/golib/util \
      gopkg.in/robfig/cron.v2 \
      gopkg.in/yaml.v2 \
      io/ioutil \
      log \
      net/http \
      path/filepath \
      time

# ============================================================
# source container contains the source as it exists within the
# repository.
FROM build as source
WORKDIR /go/src
ADD . .

# ============================================================
# Compile the source.
FROM source as compiler
ARG arch
ARG goos
ARG goarch
ARG goarm

# Microservice version is the commit hash from git
RUN version="$(git rev-parse --short HEAD)" &&\
    sed -i "s/@@version@@/${version} ${goos}(${arch})/g" bin/version.go

# Build the microservice.
# NB: CGO_ENABLED=0 forces a static build
RUN CGO_ENABLED=0 \
    GOOS=${goos} \
    GOARCH=${goarch} \
    GOARM=${goarm} \
    go build \
      -o /dest/nrod-cif \
      bin

# ============================================================
# Finally build the final runtime container for the specific
# microservice
FROM scratch

# The default database directory
Volume /database

# Install our built image
COPY --from=compiler /dest/ /

ENTRYPOINT ["/nrof-cif"]
CMD [ "-c", "/config.yaml"]
