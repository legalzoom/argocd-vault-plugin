GOARCH = amd64
UNAME = $(shell uname -s)
OS = linux
ifndef OS
	ifeq ($(UNAME), Linux)
		OS = linux
	else ifeq ($(UNAME), Darwin)
		OS = darwin
	endif
endif

.DEFAULT_GOAL := all

all: fmt build start

build:
	GOOS=$(OS) GOARCH="$(GOARCH)" CGO_ENABLED=0 go build -o vault/plugins/argocd cmd/argocd-vault-plugin/main.go

start:
	vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins

enable:
	vault secrets enable -path=argocd argocd

clean:
	rm -f ./vault/plugins/vault-plugin-secrets-mock

fmt:
	go fmt $$(go list ./...)

.PHONY: build clean fmt start enable
