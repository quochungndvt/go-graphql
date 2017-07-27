package main

import (
	"encoding/json"
	"go-graphql/data"
	"io/ioutil"
	"log"
	"os"

	"github.com/graphql-go/graphql"
)

func main() {
	// Save JSON of full schema introspection for Babel Relay Plugin to use
	result := graphql.Do(graphql.Params{
		Schema:        data.Schema,
		RequestString: data.IntrospectionQuery,
	})
	if result.HasErrors() {
		log.Fatalf("ERROR introspecting schema: %v", result.Errors)
		return
	} else {
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		err = ioutil.WriteFile("../data/schema.json", b, os.ModePerm)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}

	}
	// TODO: Save user readable type system shorthand of schema
	// pending implementation of printSchema
	/*
		fs.writeFileSync(
		  path.join(__dirname, '../data/schema.graphql'),
		  printSchema(Schema)
		);
	*/
}
