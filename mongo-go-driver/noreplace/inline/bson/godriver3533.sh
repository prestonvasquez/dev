#!bin/bash
set -ex

# compile the benchmark test
go test -c -o bench.test

# Profile BSON v1
./bench.test \
  -test.bench='^BenchmarkBSONv1vs2Comparison/BSON_v1$' \
  -test.cpuprofile=cpu_v1.prof \
  -test.benchmem \
  -test.count=10 >v1.txt

# Profile BSON v2
./bench.test \
  -test.bench='^BenchmarkBSONv1vs2Comparison/BSON_v2$' \
  -test.cpuprofile=cpu_v2.prof \
  -test.benchmem \
  -test.count=10 >v2.txt

benchstat v1.txt v2.txt

## Use go tools to open browser with comparison
#go tool pprof \
#  -http="localhost:8080" \
#  -diff_base=cpu_v1.prof \
#  bench.test cpu_v2.prof
