package main

import (
	"fmt"

	"github.com/N0el4kLs/cyberetrieve"
	"github.com/N0el4kLs/cyberetrieve/sources"
)

func main() {
	// Query struct for the search
	query := sources.Query{
		Query: `title:"login"`,
		//FofaQuery: "xxx" // specific query for fofa
		//QuakeQuery: "xxxx" //specific query for quake
		NumberOfQuery: 5, // number of query in each engine
	}
	session := sources.Session{
		QuakeToken: "xxx-xxx-xxx", // quake token
		FofaKey:    "xxx-xxx-xxx", // fofa key
	}

	// init the engine
	engine := cyberetrieve.NewCyberRetrieveEngine(query, session,
		cyberetrieve.WithAutoGrammar(), // enable auto grammar
		cyberetrieve.WithFofaSearch(),  // enable fofa search
		//cyberetrieve.WithQuakeSearch(), // enable quake search
	)

	if rst, err := engine.RetrieveResult(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(rst)
	}
}
