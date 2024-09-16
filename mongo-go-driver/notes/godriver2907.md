DRIVERS-2581 makes the definition of `mongocrypt_binary_t` public, meaning we can initialize and assign the data and length directly:

```go
mongocryptBinary := C.mongocrypt_binary_new()
if mongocryptBinary == nil {
	return nil
}

addr := (*C.uint8_t)(C.CBytes(data))

mongocryptBinary.data = unsafe.Pointer(addr)
mongocryptBinary.len = C.uint32_t(len(data)) // uint32_t
```

And access the underlying bytes:

```go  
C.GoBytes(unsafe.Pointer(b.wrapped.data), C.int(b.wrapped.len))
```

Profiling memory for this solution over a large benchtime results in a minor performance hit:

```go 
go test -bench=BenchmarkBulkDecryption -v -failfast -tags=cse -memprofile mem.prof -benchtime=2s
```

Initializing directly:
```
goos: darwin
goarch: arm64
pkg: go.mongodb.org/mongo-driver/v2/x/mongo/driver/mongocrypt
BenchmarkBulkDecryption
BenchmarkBulkDecryption/1_threads
BenchmarkBulkDecryption/1_threads-10                   2        1000009021 ns/op        14123028 B/op     588459 allocs/op
testing: BenchmarkBulkDecryption/1_threads-10 left GOMAXPROCS set to 1
BenchmarkBulkDecryption/2_threads
BenchmarkBulkDecryption/2_threads-10                   3         666674639 ns/op        13286706 B/op     553604 allocs/op
testing: BenchmarkBulkDecryption/2_threads-10 left GOMAXPROCS set to 2
BenchmarkBulkDecryption/8_threads
BenchmarkBulkDecryption/8_threads-10                  13         153852337 ns/op         5019820 B/op     209133 allocs/op
testing: BenchmarkBulkDecryption/8_threads-10 left GOMAXPROCS set to 8
BenchmarkBulkDecryption/64_threads
BenchmarkBulkDecryption/64_threads-10                 70          28573686 ns/op          863848 B/op      35856 allocs/op
testing: BenchmarkBulkDecryption/64_threads-10 left GOMAXPROCS set to 64
PASS
ok      go.mongodb.org/mongo-driver/v2/x/mongo/driver/mongocrypt        22.759s
```

Original:
```
goos: darwin
goarch: arm64
pkg: go.mongodb.org/mongo-driver/v2/x/mongo/driver/mongocrypt
BenchmarkBulkDecryption
BenchmarkBulkDecryption/1_threads
BenchmarkBulkDecryption/1_threads-10                   2        1000007167 ns/op        13623020 B/op     567625 allocs/op
testing: BenchmarkBulkDecryption/1_threads-10 left GOMAXPROCS set to 1
BenchmarkBulkDecryption/2_threads
BenchmarkBulkDecryption/2_threads-10                   3         666673444 ns/op        12745938 B/op     531072 allocs/op
testing: BenchmarkBulkDecryption/2_threads-10 left GOMAXPROCS set to 2
BenchmarkBulkDecryption/8_threads
BenchmarkBulkDecryption/8_threads-10                  13         153852109 ns/op         4725225 B/op     196844 allocs/op
testing: BenchmarkBulkDecryption/8_threads-10 left GOMAXPROCS set to 8
BenchmarkBulkDecryption/64_threads
BenchmarkBulkDecryption/64_threads-10                 70          28574396 ns/op          814259 B/op      33806 allocs/op
testing: BenchmarkBulkDecryption/64_threads-10 left GOMAXPROCS set to 64
PASS
ok      go.mongodb.org/mongo-driver/v2/x/mongo/driver/mongocrypt        22.885s
```

There does not appear to be extra cost to calling C from Go.
