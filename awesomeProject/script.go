package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"io/ioutil"
)

func LoadJson(file string) []byte {
	var data []byte
	data, _ = ioutil.ReadFile(file)
	fmt.Println(data)
	return data
}

func main() {
	var jsonData []byte
	jsonData = LoadJson("model.json")
	sc, _ := stan.Connect("test-cluster", "writer")
	sc.Publish("foo", jsonData)
	sc.Close()
}
