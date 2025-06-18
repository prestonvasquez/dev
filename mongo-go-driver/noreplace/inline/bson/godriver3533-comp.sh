#!/bin/bash

set +ex

# compile the benchmark test
go test -c -o bench.test

go tool pprof \
  -diff_base=cpu_v1.prof \
  -text \
  -drop_negative \
  -flat \
  -nodecount=20 \
  bench.test cpu_v2.prof
