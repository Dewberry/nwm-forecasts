FROM golang:1.14 as dev

WORKDIR /fetch
COPY ./ /fetch

# install the netCDF C package dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libnetcdf-dev \
    ca-certificates \
    curl \
    fuse \
    && rm -rf /var/lib/apt/lists/* \
    && curl -o /usr/local/bin/goofys -J -L -H "Accept: application/octet-stream" https://github.com/kahing/goofys/releases/download/v0.24.0/goofys \
    && curl -o /usr/local/bin/catfs -J -L -H "Accept: application/octet-stream" https://github.com/kahing/catfs/releases/download/v0.8.0/catfs \
    && chmod +x /usr/local/bin/goofys \
    && chmod +x /usr/local/bin/catfs \
    && apt-get purge -y --auto-remove \
    curl

ENV GOBIN /go/bin
RUN go mod download

# Hot-Reloader
RUN go get github.com/githubnemo/CompileDaemon

# This is not required for the dev build, but needed for prod
RUN go build main.go 

ENTRYPOINT CompileDaemon --build="go build main.go" --command="./main"

#-------------PRODUCTION IMAGE---------------#
FROM buildpack-deps:buster-scm as prod

# install the netCDF C package
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libnetcdf-dev \
    ca-certificates \
    curl \
    fuse \
    && rm -rf /var/lib/apt/lists/* \
    && curl -o /usr/local/bin/goofys -J -L -H "Accept: application/octet-stream" https://github.com/kahing/goofys/releases/download/v0.24.0/goofys \
    && curl -o /usr/local/bin/catfs -J -L -H "Accept: application/octet-stream" https://github.com/kahing/catfs/releases/download/v0.8.0/catfs \
    && chmod +x /usr/local/bin/goofys \
    && chmod +x /usr/local/bin/catfs \
    && apt-get purge -y --auto-remove \
    curl

COPY --from=dev /fetch/main /

RUN mkdir /retrospective
RUN mkdir /forecast

ENTRYPOINT [ "/main" ]