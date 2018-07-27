# Inspec exporter

Inspec exporter for complaince metrics, written
in Go with pluggable metric collectors. Its possible to gather results local or remote over ssh.

## Building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/)
* RHEL/CentOS: `glibc-static` package.

Building:

    go get gitlab.tsi-dev.otc-service.com/tsi-automation/inspec_exporter
    cd ${GOPATH-$HOME/go}/src/gitlab.tsi-dev.otc-service.com/tsi-automation/inspec_exporter
    make
    ./inspec_exporter <flags>

To see all available configuration flags:

    ./inspec_exporter -h

## Running tests

    make test


## Using Docker

TBD

## Inspec Config

TBD

## Remote exec

TBD