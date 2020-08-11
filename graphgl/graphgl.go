package graphgl

import (
	"context"
	"net/http"

	enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	"github.com/friendsofgo/graphiql"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

//GraphQL
const Schema = `
type Customer_struct {
    Customer_id: String!
    Customer_name: String!
	Customer_type: String!
	Customer_email: String!
}
type Query {
    FindOneRow(Customer_id: String!): Customer_struct
}
schema {
    query: Query
}
`

type FindOneRow_Resolver struct {
	v *rootsctuct.Customer_struct
}

func (r *FindOneRow_Resolver) Customer_id() string    { return r.v.Customer_id }
func (r *FindOneRow_Resolver) Customer_name() string  { return r.v.Customer_name }
func (r *FindOneRow_Resolver) Customer_type() string  { return r.v.Customer_type }
func (r *FindOneRow_Resolver) Customer_email() string { return r.v.Customer_email }

func (q *query) FindOneRow(ctx context.Context, args struct{ Customer_id string }) *FindOneRow_Resolver {

	v, err := enginecrm.EngineCRMv.FindOneRow(enginecrm.EngineCRMv.DataBaseType, args.Customer_id, rootsctuct.Global_settingsV)

	if err != nil {
		enginecrm.EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil
	}

	return &FindOneRow_Resolver{v: &v}
	//return &v
}

type query struct{}

//GraphQL end

func StartGraphQL() {

	//GraphQL
	httpGraphQL := http.NewServeMux()

	schema := graphql.MustParseSchema(Schema, &query{})
	httpGraphQL.Handle("/query", &relay.Handler{Schema: schema})

	// First argument must be same as graphql handler path
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/query")
	if err != nil {
		panic(err)
	}
	httpGraphQL.Handle("/", graphiqlHandler)

	go http.ListenAndServe(":8184", httpGraphQL)
	//GraphQL end

}
