FROM golang:1.13 as build

ENV GO111MODULE=on

ADD tools tools
RUN cd tools/health && go build -o ../../health main.go

RUN useradd -u 1001 nifikop
RUN mkdir -p /tmp && chown nifikop /tmp

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

LABEL org.opencontainers.image.documentation="https://github.com/Orange-OpenSource/nifikop/blob/master/README.md"
LABEL org.opencontainers.image.authors="Alexandre Guitton <aguitton.ext@orange.com>"
LABEL org.opencontainers.image.source="https://github.com/Orange-OpenSource/nifikop"
LABEL org.opencontainers.image.vendor="Orange France - Digital Factory"
LABEL org.opencontainers.image.version="0.1"
LABEL org.opencontainers.image.description="Operateur des Gestion de Clusters Nifi"
LABEL org.opencontainers.image.url="https://github.com/Orange-OpenSource/nifikop"
LABEL org.opencontainers.image.title="Operateur NiFi"

LABEL org.label-schema.usage="https://github.com/Orange-OpenSource/nifikop/blob/master/README.md"
LABEL org.label-schema.docker.cmd="/usr/local/bin/nifikop"
LABEL org.label-schema.docker.cmd.devel="N/A"
LABEL org.label-schema.docker.cmd.test="N/A"
LABEL org.label-schema.docker.cmd.help="N/A"
LABEL org.label-schema.docker.cmd.debug="N/A"
LABEL org.label-schema.docker.params="LOG_LEVEL=define loglevel,RESYNC_PERIOD=period in second to execute resynchronisation,WATCH_NAMESPACE=namespace to watch for nificlusters,OPERATOR_NAME=name of the operator instance pod"

ENV OPERATOR=/usr/local/bin/nifikop \
    USER_UID=1001 \
    USER_NAME=nifikop

COPY --from=build /tmp /tmp
COPY --from=build /go/health /

# install operator binary
COPY build/_output/bin/nifikop ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
