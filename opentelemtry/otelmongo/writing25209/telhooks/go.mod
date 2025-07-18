module github.com/prestonvasquez/dev/opentelemetry/otelmongo/writing25209/telhooks

go 1.23.1

replace go.mongodb.org/mongo-driver/v2 => /Users/preston.vasquez/Developer/mongo-go-driver

replace go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo => /Users/preston.vasquez/Developer/opentelemetry-go-contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo

replace go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo => /Users/preston.vasquez/Developer/opentelemetry-go-contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo

require (
	go.mongodb.org/mongo-driver/v2 v2.0.0
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
