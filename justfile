# BLACKSWAN - Investment Microservice Monorepo
# Justfile for build automation and development tasks

# Variables
GO := "go"
GO_TEST := "gotestsum --"
GO_BUILD := GO + " build"
GO_VET := GO + " vet"
GO_CLEAN := GO + " clean"
GO_GET := GO + " get"
GO_LINT := "golangci-lint"

# Build directories
BUILD_DIR := "./bin"
DATA_DIR := "./data"

# Default list command
default: list

# Display all available tasks
list:
    @just --list

# Build all binaries
build:

#
lint:

# 
test:
