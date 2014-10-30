FROM ubuntu:14.04
MAINTAINER rafael.colton@gmail.com

ENV DEBIAN_FRONTEND noninteractive
ENV GOROOT /usr/local/go
ENV GOPATH /app
ENV GO_TARBALL go1.3.1.linux-amd64.tar.gz
ENV LD_LIBRARY_PATH /lib/x86_64-linux-gnu:/usr/local/lib:/usr/lib:/lib

# Fix some issues with APT packages
# See https://github.com/dotcloud/docker/issues/1024
RUN dpkg-divert --local --rename --add /sbin/initctl
RUN ln -sFf /bin/true /sbin/initctl
RUN echo "initscripts hold" | dpkg --set-selections

# install deps
RUN apt-get update -y && apt-get install -y -qq --no-install-recommends apt-transport-https \
  build-essential curl openssh-client make git-core pkg-config mercurial ca-certificates

# install go
RUN curl -sLO https://storage.googleapis.com/golang/$GO_TARBALL
RUN tar -C /usr/local -xzf $GO_TARBALL
RUN ln -sv /usr/local/go/bin/* /usr/local/bin
RUN rm -f $GO_TARBALL

WORKDIR /app/src/github.com/rafecolton/docker-builder

# set up build dir and add project
ADD . /app/src/github.com/rafecolton/docker-builder

# make sure we don't have trouble getting deps from GitHub
RUN ssh-keyscan github.com > /etc/ssh/ssh_known_hosts

# install and verify
RUN touch Makefile
RUN make build

CMD ["-h"]
ENTRYPOINT ["/app/bin/docker-builder"]
