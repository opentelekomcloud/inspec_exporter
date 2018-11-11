# Inspec exporter

[![Build Status](https://travis-ci.org/OpenTelekomCloud/inspec_exporter.svg?branch=master)](https://travis-ci.org/OpenTelekomCloud/inspec_exporter)

Inspec exporter for complaince metrics, written
in Go with pluggable metric collectors. Its possible to gather results local or remote over ssh.

## Building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/)
* RHEL/CentOS: `glibc-static` package.

Building:

    go get github.com/OpenTelekomCloud/inspec_exporter
    cd $GOPATH/src/github.com/OpenTelekomCloud/inspec_exporter
    make
    ./inspec_exporter <flags>

To see all available configuration flags:

    ./inspec_exporter -h

## Running tests

    make test


## Using Docker

    make docker
    docker run --name inspec_exporter -p 9124:9124 -v /srv/insepc:/profiles:ro -v /etc/inspec_exporter/inspec.yml:/etc/inspec_exporter/inspec.yml:ro -d inspec_exporter:master

## Inspec Config

TBD

## Remote exec

TBD