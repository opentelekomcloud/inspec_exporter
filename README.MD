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

TBD

## Inspec Config

TBD

## Remote exec

TBD