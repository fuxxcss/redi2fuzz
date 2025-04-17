#!/bin/bash

go build -o build/libmutator.so -buildmode=c-shared pkg/mutator.go
go build -o build/driver pkg/driver.go
go build

