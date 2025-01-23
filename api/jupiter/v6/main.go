package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {

	url := "https://quote-api.jup.ag/v6/swap"
	method := "POST"

	payload := strings.NewReader(`{
  "userPublicKey": "validPublicKey",
  "wrapAndUnwrapSol": true,
  "quoteResponse": {
    "inputMint": "validMintAddress",
    "inAmount": "1000000",
    "outputMint": "validMintAddress",
    "swapMode": "ExactIn"
  }
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Fatalf("failed to create http request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to make http request: %v", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
	}

	fmt.Println("response: ", string(body))
}
