package gqlrelay_test

import (
	"fmt"
	"github.com/chris-ramon/graphql"
	"github.com/chris-ramon/graphql/testutil"
	"github.com/sogko/graphql-relay-go"
	"reflect"
	"testing"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type photo struct {
	ID    int `json:"id"`
	Width int `json:"width"`
}

var nodeTestUserData = map[string]*user{
	"1": &user{1, "John Doe"},
	"2": &user{2, "Jane Smith"},
}
var nodeTestPhotoData = map[string]*photo{
	"3": &photo{3, 300},
	"4": &photo{4, 400},
}

// declare types first, define later in init()
// because they all depend on nodeTestDef
var nodeTestUserType *graphql.Object
var nodeTestPhotoType *graphql.Object

var nodeTestDef = gqlrelay.NewNodeDefinitions(gqlrelay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo) interface{} {
		if user, ok := nodeTestUserData[id]; ok {
			return user
		}
		if photo, ok := nodeTestPhotoData[id]; ok {
			return photo
		}
		return nil
	},
	TypeResolve: func(value interface{}, info graphql.ResolveInfo) *graphql.Object {
		switch value.(type) {
		case *user:
			return nodeTestUserType
		case *photo:
			return nodeTestPhotoType
		default:
			panic(fmt.Sprintf("Unknown object type `%v`", value))
		}
	},
})
var nodeTestQueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.FieldConfigMap{
		"node": nodeTestDef.NodeField,
	},
})

// becareful not to define schema here, since nodeTestUserType and nodeTestPhotoType wouldn't be defined till init()
var nodeTestSchema graphql.Schema

func init() {
	nodeTestUserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.FieldConfigMap{
			"id": &graphql.FieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"name": &graphql.FieldConfig{
				Type: graphql.String,
			},
		},
		Interfaces: []*graphql.Interface{nodeTestDef.NodeInterface},
	})
	nodeTestPhotoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.FieldConfigMap{
			"id": &graphql.FieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"width": &graphql.FieldConfig{
				Type: graphql.Int,
			},
		},
		Interfaces: []*graphql.Interface{nodeTestDef.NodeInterface},
	})

	nodeTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: nodeTestQueryType,
	})
}

func testGraphql(t *testing.T, p graphql.Params) *graphql.Result {
	resultChannel := make(chan *graphql.Result)
	go graphql.Graphql(p, resultChannel)
	result := <-resultChannel
	return result
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectIDForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "1",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectIDForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "4",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectNameForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          ... on User {
            name
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":   "1",
				"name": "John Doe",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectWidthForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
          ... on Photo {
            width
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":    "4",
				"width": 400,
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectTypeNameForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          __typename
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":         "1",
				"__typename": "User",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectTypeNameForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
          __typename
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":         "4",
				"__typename": "Photo",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_IgnoresPhotoFragmentsOnUser(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          ... on Photo {
            width
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "1",
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_ReturnsNullForBadIDs(t *testing.T) {
	query := `{
        node(id: "5") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": nil,
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_CorrectlyIntrospects_HasCorrectNodeInterface(t *testing.T) {
	query := `{
        __type(name: "Node") {
          name
          kind
          fields {
            name
            type {
              kind
              ofType {
                name
                kind
              }
            }
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "Node",
				"kind": "INTERFACE",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "id",
						"type": map[string]interface{}{
							"kind": "NON_NULL",
							"ofType": map[string]interface{}{
								"name": "ID",
								"kind": "SCALAR",
							},
						},
					},
				},
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_CorrectlyIntrospects_HasCorrectNodeRootField(t *testing.T) {
	query := `{
        __schema {
          queryType {
            fields {
              name
              type {
                name
                kind
              }
              args {
                name
                type {
                  kind
                  ofType {
                    name
                    kind
                  }
                }
              }
            }
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "node",
							"type": map[string]interface{}{
								"name": "Node",
								"kind": "INTERFACE",
							},
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"name": "ID",
											"kind": "SCALAR",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	result := testGraphql(t, graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
