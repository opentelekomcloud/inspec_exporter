FROM        ruby:alpine
MAINTAINER  Sascha Veres <sascha.veres@t-systems.com>

ARG INSPEC_VERSION=2.2.41
ARG GEM_SOURCE=https://rubygems.org

COPY inspec_exporter  /bin/inspec_exporter
COPY inspec.yml       /etc/inspec_exporter/inspec.yml

RUN mkdir -p /share
RUN apk add --update build-base libxml2-dev libffi-dev git && \
    gem install --no-document --source ${GEM_SOURCE} --version ${INSPEC_VERSION} inspec && \
    apk del build-base

EXPOSE      9124
ENTRYPOINT  [ "/bin/inspec_exporter" ]
CMD         [ "--config.file=/etc/inspec_exporter/inspec.yml" ]

VOLUME ["/profiles"]
WORKDIR /etc/inspec_exporter/profiles