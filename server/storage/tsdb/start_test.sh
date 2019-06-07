#!/bin/bash

go test -c -o raw_tsdb_test raw_tsdb_test.go
./raw_tsdb_test -test.v -test.count 100
