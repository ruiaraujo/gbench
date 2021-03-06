package main

var Schema = `
	schema {
		query: Query
	}

	interface Base {
		id: String
	}

	type Query implements Base {
		id: String
		string: String
		listOfStrings: [String]
		listOfObjects: [Query]
		listOfInterfaces: [Base]
	}
`

type base interface {
	ID()	*string
}

type query struct {
	base
	ID        string
	String      string
	// ListOfStrings   []string
	// ListOfObjects []query
}

type Resolver struct{
	q *query
}

var listOfObjects []*Resolver;
var listOfStrings []*string;
var listOfInterfaces []*Resolver;

var baseResolver Resolver;

func init() {
	var helloWorld string
	helloWorld = "Hello World!"
	// var ob query
	// ob = query {
	// 	ID: "id",
	// 	String: "Hello World!",
	// }
	baseResolver = Resolver{
		q: &query {
			ID: "id",
			String: "Hello World!",
		},
	}
	listOfStrings = make([]*string, 100)
	listOfObjects = make([]*Resolver, 100)
	listOfInterfaces = make([]*Resolver, 100)

	for i := 0; i < 100; i++ {
		listOfStrings[i] = &helloWorld
		listOfObjects[i] = &baseResolver
		listOfInterfaces[i] = &baseResolver
	}
}

func (r *Resolver) ID() *string {
	return &r.q.ID
}

func (r *Resolver) String() *string {
	return &r.q.String
}

func (r *Resolver) ListOfStrings() *[]*string {
	return &listOfStrings
}

func (r *Resolver) ListOfObjects() *[]*Resolver {
	return &listOfObjects
}

func (r *Resolver) ListOfInterfaces() *[]*Resolver {
	return &listOfInterfaces
}

func (r *Resolver) ToQuery() (*Resolver, bool) {
	return &baseResolver, true
}
