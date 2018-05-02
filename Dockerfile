FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Sascha Veres <sascha.veres@t-systems.com>

COPY inspec_exporter  /bin/inspec_exporter
COPY inspec.yml       /etc/inspec_exporter/inspec.yml

EXPOSE      9124
ENTRYPOINT  [ "/bin/inspec_exporter" ]
CMD [ "--config.file=/etc/inspec_exporter/inspec.yml" ]