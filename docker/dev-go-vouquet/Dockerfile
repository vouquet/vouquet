FROM ubuntu:20.04

ENV GOVERSION 1.15.8

ENV GOROOT /opt/go
ENV GOPATH /go/src

ENV LANG en_US.UTF-8
#ENV http_proxy http://proxy.i.hinoshiba.com:8080/
#ENV https_proxy http://proxy.i.hinoshiba.com:8080/
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && \
    apt install -y git curl wget tzdata make gcc && \
    apt clean && \
    rm -rf /var/lib/apt/lists/* && \
    ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

RUN cd /opt && wget -nv --show-progress --progress=bar:force:noscroll https://golang.org/dl/go${GOVERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOVERSION}.linux-amd64.tar.gz && rm go${GOVERSION}.linux-amd64.tar.gz && \
    wget -nv --show-progress --progress=bar:force:noscroll https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 -O /opt/go/bin/dep && \
    chmod 755 /opt/go/bin/dep && \
    ln -s /opt/go/bin/go /usr/bin/ && \
    ln -s /opt/go/bin/dep /usr/bin/ && \
    mkdir -p ${GOPATH} && \
    mkdir /.cache && \
    chmod 777 /.cache

WORKDIR ${GOPATH}
