FROM centos:latest 
MAINTAINER Cosmin Ilie <cosmin_ilie@live.com>

ENV CI_PROJECT_DIR /builds/cosminilie/gitbot 
RUN yum install -y epel-release

RUN yum update -y \
 && yum-config-manager --disable extras \
 && yum makecache \
 && yum install -y rpm-build rpmdevtools yum-utils sudo tar ccache make rpmlint gcc git wget\
 && yum clean packages #20161006 \
 && sed -i 's/\(keepcache=\)0/\11/' /etc/yum.conf && mkdir -p /builds/cosminilie/gitbot

ENV GOLANG_VERSION  1.7.5
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 2e4dd6c44f0693bef4e7b46cc701513d74c3cc44f2419bf519d7868b12931ac3

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz


ENV GOPATH /tmp/gopath
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"


COPY . /builds/cosminilie/gitbot


