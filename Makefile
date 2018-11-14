BASIC=Makefile
GO=go
GIT=git

all: .PHONY ingest

ingest: .PHONY
	$(GO) build $(LD_FLAGS)

fmt: .PHONY
	$(GO) fmt

vet: .PHONY
	$(GO) vet

test: .PHONY vet
	$(GO) test

.PHONY:
