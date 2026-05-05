package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var svcConfig tSvcConfig

func LoadXMLFromFile(filename string, v interface{}) {
	xmlFile, err := os.Open(filename)
	if err != nil {
		log.Println(err)
	}
	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(byteValue, v)
}

var mapOfDBPools map[string]*pgxpool.Pool

func main() {
	LoadXMLFromFile("preferences.xml", &svcConfig)
	var err error
	mapOfDBPools = make(map[string]*pgxpool.Pool)
	for _, c := range svcConfig.Connections.Connection {
		mapOfDBPools[c.Name], err = pgxpool.New(context.Background(), c.DSN)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			os.Exit(1)
		}
		defer mapOfDBPools[c.Name].Close()
	}

	fmt.Println("Checking DB connections: ", checkDBConnections())
	//	testStructure()
	//	token, _ := getToken("123")
	//err = updateToken(token)
	//	fmt.Println(XML2String(token))
	startSVC()
}
