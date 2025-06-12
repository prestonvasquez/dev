module github.com/prestonvasquez/dev/mongo-go-driver/noreplace

go 1.23.1

replace go.mongodb.org/mongo-driver/v2 => /Users/preston.vasquez/Developer/mongo-go-driver

require (
	github.com/stretchr/testify v0.0.0-20161117074351-18a02ba4a312
	go.mongodb.org/mongo-driver v1.17.3
	go.mongodb.org/mongo-driver/v2 v2.2.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v0.0.0-20151028094244-d8ed2627bdf0 // indirect
)
