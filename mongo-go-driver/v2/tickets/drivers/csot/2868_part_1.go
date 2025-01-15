package main

// If maxAwaitTimeMS is set to 9ms, timeoutMS to 10ms, and RTT is 2ms, this
// configuration might result in getMore command timing out when no data is
// available, leading to connection closure on the driver's side
func main() {

}
