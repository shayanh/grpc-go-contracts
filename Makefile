SHELL=/bin/bash

GOFILES_NOVENDOR = $(shell go list ./... | grep -v /vendor/)

all: vet fmt

fmt:
	go fmt $(GOFILES_NOVENDOR)

vet:
	go vet $(GOFILES_NOVENDOR)

.PHONY: all fmt vet