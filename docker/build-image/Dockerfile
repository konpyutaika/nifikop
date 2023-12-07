ARG GOLANG_VERSION

FROM docker as docker-base
COPY --from=docker/buildx-bin:latest /buildx /usr/libexec/docker/cli-plugins/docker-buildx

#https://github.com/CircleCI-Public/circleci-dockerfiles/blob/master/golang/images/1.12.4/Dockerfile
FROM golang:${GOLANG_VERSION}

# Make Apt non-interactive
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/90circleci \
  && echo 'DPkg::Options "--force-confnew";' >> /etc/apt/apt.conf.d/90circleci

ENV DEBIAN_FRONTEND=noninteractive

# Debian Jessie is EOL'd and original repos don't work.
# Switch to the archive mirror until we can get people to
# switch to Stretch.
RUN if grep -q Debian /etc/os-release && grep -q jessie /etc/os-release; then \
  rm /etc/apt/sources.list \
  && echo "deb http://archive.debian.org/debian/ jessie main" >> /etc/apt/sources.list \
  && echo "deb http://security.debian.org/debian-security jessie/updates main" >> /etc/apt/sources.list \
  ; fi

# Make sure PATH includes ~/.local/bin
# https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=839155
RUN echo 'PATH="$HOME/.local/bin:$PATH"' >> /etc/profile.d/user-local-path.sh

# man directory is missing in some base images
# https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=863199
RUN apt-get update \
  && mkdir -p /usr/share/man/man1 \
  && apt-get install -y \
  git apt \
  locales sudo openssh-client ca-certificates tar gzip \
  net-tools netcat-traditional unzip zip bzip2 gnupg curl wget

# Set timezone to UTC by default
RUN ln -sf /usr/share/zoneinfo/Etc/UTC /etc/localtime

# Use unicode
RUN locale-gen C.UTF-8 || true
ENV LANG=C.UTF-8

# Install jq
RUN JQ_URL="https://circle-downloads.s3.amazonaws.com/circleci-images/cache/linux-amd64/jq-latest" \
  && curl --silent --show-error --location --fail --retry 3 --output /usr/bin/jq $JQ_URL \
  && chmod +x /usr/bin/jq \
  && jq --version

# Install Docker
RUN set -ex \
  && export DOCKER_VERSION=$(curl --silent --fail --retry 3 https://download.docker.com/linux/static/stable/x86_64/ | grep -o -e 'docker-[.0-9]*\.tgz' | sort -r | head -n 1) \
  && DOCKER_URL="https://download.docker.com/linux/static/stable/x86_64/${DOCKER_VERSION}" \
  && echo Docker URL: $DOCKER_URL \
  && curl --silent --show-error --location --fail --retry 3 --output /tmp/docker.tgz "${DOCKER_URL}" \
  && ls -lha /tmp/docker.tgz \
  && tar -xz -C /tmp -f /tmp/docker.tgz \
  && mv /tmp/docker/* /usr/bin \
  && rm -rf /tmp/docker /tmp/docker.tgz \
  && which docker \
  && (docker version || true)

# Install docker-buildx plugin from docker/buildx-bin image
COPY --from=docker-base /usr/libexec/docker/cli-plugins/docker-buildx /usr/libexec/docker/cli-plugins/docker-buildx

RUN groupadd --gid 3434 circleci \
  && useradd --uid 3434 --gid circleci --shell /bin/bash --create-home circleci \
  && echo 'circleci ALL=NOPASSWD: ALL' >> /etc/sudoers.d/50-circleci \
  && echo 'Defaults    env_keep += "DEBIAN_FRONTEND"' >> /etc/sudoers.d/env_keep

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | INSTALL_DIRECTORY=/usr/local/bin sh
RUN curl -sSL https://github.com/gotestyourself/gotestsum/releases/download/v0.3.4/gotestsum_0.3.4_linux_amd64.tar.gz | \
  tar -xz -C /usr/local/bin gotestsum

# Install official golang CI linter
ARG GOLANGCI_VERSION
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v${GOLANGCI_VERSION}

# Install kubectl cli
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl  \
  && chmod +x /usr/local/bin/kubectl

# Install helm client
RUN curl -L https://get.helm.sh/helm-v3.8.0-linux-amd64.tar.gz -o helm.tar.gz \
  && tar zxf helm.tar.gz \
  && mv linux-amd64/helm /usr/local/bin \
  && rm -rf helm.tar.gz linux-amd64

WORKDIR /go/nifikop

ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

CMD ["/bin/sh"]