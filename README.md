# Go Insteon

[![Go Report Card](https://goreportcard.com/badge/github.com/abates/insteon)](https://goreportcard.com/report/github.com/abates/insteon) [![Build Status](https://travis-ci.org/abates/insteon.svg?branch=master)](https://travis-ci.org/abates/insteon) [![GoDoc](https://godoc.org/github.com/abates/insteon?status.png)](https://godoc.org/github.com/abates/insteon) [![Coverage Status](https://coveralls.io/repos/github/abates/insteon/badge.svg?branch=master)](https://coveralls.io/github/abates/insteon?branch=master)

This package provides a Go interface to Insteon networks and the ability to
control Insteon devices. This package is being actively developed and the
API is subject to change. Consider this library to be in an alpha stage of
development

## CLI Utility

The package provides the "ic" command line tool to perform various
administrative tasks related to the Insteon network and its devices.

## Insteon Network Daemon
TODO: A REST interface to the Insteon network. Will include abstractions for
common tasks such as creating virtual N-Way light switches as well as scenes

## Insteon Network Client
TODO: A client application to the Insteon Network Daemon

## API

The package can be used directly from other go programs by means of the
github.com/abates/insteon package.  See the
[godocs](https://godoc.org/github.com/abates/insteon) for more information.
