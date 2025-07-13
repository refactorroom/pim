package main

import pim "github.com/refactorroom/pim"

// Example_json demonstrates JSON logging with pim.
func Example_json() {
	data := map[string]interface{}{"user": "alice", "active": true}
	pim.Json(data)
}
