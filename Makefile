.PHONY: build
.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o calendar .
