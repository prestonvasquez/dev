package main

import (
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
)

// redactedCmd is mongodb server response message.
const redactedCmd = `DQIAABBuAGQAAAAEdXBzZXJ0ZWQA3AEAAAMwACEAAAAQaW5kZXgABQAAAAdfaWQA
aGN3+oAWCmDIfXkPAAMxACEAAAAQaW5kZXgABwAAAAdfaWQAaGN3+oAWCmDIfXkQ
AAMyACEAAAAQaW5kZXgAFAAAAAdfaWQAaGN3+oAWCmDIfXkRAAMzACEAAAAQaW5k
ZXgAHAAAAAdfaWQAaGN3+oAWCmDIfXkSAAM0ACEAAAAQaW5kZXgAIwAAAAdfaWQA
aGN3+oAWCmDIfXkTAAM1ACEAAAAQaW5kZXgAJQAAAAdfaWQAaGN3+oAWCmDIfXkU
AAM2ACEAAAAQaW5kZXgAPgAAAAdfaWQAaGN3+oAWCmDIfXkVAAM3ACEAAAAQaW5k
ZXgASgAAAAdfaWQAaGN3+oAWCmDIfXkWAAM4ACEAAAAQaW5kZXgAUQAAAAdfaWQA
aGN3+oAWCmDIfXkXAAM5ACEAAAAQaW5kZXgAVwAAAAdfaWQAaGN3+oAWCmDIfXkY
AAMxMAAhAAAAEGluZGV4AFwAAAAHX2lkAGhjd/qAFgpgyH15GQADMTEAIQAAABBp
bmRleABdAAAAB19pZABoY3f6gBYKYMh9eRoAAzEyACEAAAAQaW5kZXgAYwAAAAdf
aWQAaGN3+oAWCmDIfXkbAAAQbk1vZGlmaWVkAFcAAAABb2sAAAAAAAAA8D8A`

func main() {
	raw, _ := base64.StdEncoding.DecodeString(redactedCmd)
	str := bsoncore.Document(raw).StringN(1000) // panic: runtime error: slice bounds out of range [:-1]
	fmt.Println(str, len(str))
}
