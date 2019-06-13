#!/bin/bash

go test -c -o tsdb_test .
./tsdb_test -test.v -test.count 100
