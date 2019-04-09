package main

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"log"
	"net/http"
)

type params struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func executeQuery(p params, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{

		Schema:         schema,
		RequestString:  p.Query,
		OperationName:  p.OperationName,
		VariableValues: p.Variables,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

func main() {
	stringList := make([]string, 100)
	for i := 0; i < 100; i++ {
		stringList[i] = "Hello World!"
	}
	objectList := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		objectList[i] = "non null"
	}
	var rootQuery *graphql.Object
	base := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Base",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			return rootQuery
		}})
	// Schema
	fields := graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "id", nil
			},
		},
		"string": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Hello World!", nil
			},
		},
		"listOfStrings": &graphql.Field{
			Type: graphql.NewList(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return stringList, nil
			},
		},
		"listOfInterfaces": &graphql.Field{
			Type: graphql.NewList(base),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return objectList, nil
			},
		},
	}
	rootQuery = graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: fields, Interfaces: []*graphql.Interface{base}})
	fields["listOfObjects"]= &graphql.Field{
		Type: graphql.NewList(rootQuery),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return objectList, nil
		},
	}
	schemaConfig := graphql.SchemaConfig{Query: rootQuery}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		var p params
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := executeQuery(p, schema)
		json.NewEncoder(w).Encode(result)
	})

	fmt.Println("Now server is running on port 8083")
	err = http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatalf("failed to start new server, error: %v", err)
	}

}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/graphql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}
			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
