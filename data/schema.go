package data

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"golang.org/x/net/context"
)

var todoType *graphql.Object
var userType *graphql.Object

var nodeDefinitions *relay.NodeDefinitions
var todosConnection *relay.GraphQLConnectionDefinitions

var Schema graphql.Schema

func init() {
	nodeDefinitions = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
		IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
			resolvedID := relay.FromGlobalID(id)
			if resolvedID.Type == "Todo" {
				return GetTodo(resolvedID.ID), nil
			}
			if resolvedID.Type == "User" {
				return GetUser(resolvedID.ID), nil
			}
			return nil, nil
		},
		TypeResolve: func(p graphql.ResolveTypeParams) *graphql.Object {
			value := p.Value
			switch value.(type) {
			case *Todo:
				return todoType
			case *User:
				return userType
			default:
				return userType
			}
		},
	})
	todoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Todo",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("Todo", nil),
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"note": &graphql.Field{
				Type: graphql.String,
			},
			"complete": &graphql.Field{
				Type: graphql.Boolean,
			},
			"created_by": &graphql.Field{
				Type: graphql.String,
			},
			"created_at": &graphql.Field{
				Type: graphql.String,
			},
			"updated_at": &graphql.Field{
				Type: graphql.String,
			},
			"remind_at": &graphql.Field{
				Type: graphql.String,
			},
			"repeat_every": &graphql.Field{
				Type: graphql.String,
			},
			"important": &graphql.Field{
				Type: graphql.String,
			},
		},
		Interfaces: []*graphql.Interface{nodeDefinitions.NodeInterface},
	})

	todosConnection = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Todo",
		NodeType: todoType,
	})
	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("User", nil),
			"todos": &graphql.Field{
				Type: todosConnection.ConnectionType,
				Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
					"status": &graphql.ArgumentConfig{
						Type:         graphql.String,
						DefaultValue: "any",
					},
				}),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					status, _ := p.Args["status"].(string)
					args := relay.NewConnectionArguments(p.Args)
					todos := TodosToSliceInterface(GetTodos(status))
					return relay.ConnectionFromArray(todos, args), nil
				},
			},
			"totalCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return len(GetTodos("any")), nil
				},
			},
			"completedCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return len(GetTodos("completed")), nil
				},
			},
		},
		Interfaces: []*graphql.Interface{nodeDefinitions.NodeInterface},
	})

	rootType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Root",
		Fields: graphql.Fields{
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
			"node": nodeDefinitions.NodeField,
		},
	})

	addTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "AddTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"important": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"todoEdge": &graphql.Field{
				Type: todosConnection.EdgeType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					fmt.Println("addTodoMutation", payload)
					return relay.EdgeType{
						Node:   todo,
						Cursor: relay.CursorForObjectInConnection(TodosToSliceInterface(GetTodos("any")), todo),
					}, nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			title, _ := inputMap["title"].(string)
			important, _ := inputMap["important"].(string)
			todoId := AddTodo(title, important, false)
			return map[string]interface{}{
				"todoId": todoId,
			}, nil
		},
	})
	changeTodoStatusMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "ChangeTodoStatus",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"complete": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
		},
		OutputFields: graphql.Fields{
			"todo": &graphql.Field{
				Type: todoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					return todo, nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			id, _ := inputMap["id"].(string)
			complete, _ := inputMap["complete"].(bool)
			resolvedId := relay.FromGlobalID(id)
			ChangeTodoStatus(resolvedId.ID, complete)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}, nil
		},
	})

	markAllTodosMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "MarkAllTodos",
		InputFields: graphql.InputObjectConfigFieldMap{
			"complete": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
		},
		OutputFields: graphql.Fields{
			"changedTodos": &graphql.Field{
				Type: graphql.NewList(todoType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					todoIds, _ := payload["todoIds"].([]string)
					todos := []*Todo{}
					for _, todoId := range todoIds {
						todo := GetTodo(todoId)
						if todo != nil {
							todos = append(todos, todo)
						}
					}
					return todos, nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			complete, _ := inputMap["complete"].(bool)
			todoIds := MarkAllTodos(complete)
			return map[string]interface{}{
				"todoIds": todoIds,
			}, nil
		},
	})

	removeCompletedTodosMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RemoveCompletedTodos",
		OutputFields: graphql.Fields{
			"deletedTodoIds": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoIds"], nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			todoIds := RemoveCompletedTodos()
			return map[string]interface{}{
				"todoIds": todoIds,
			}, nil
		},
	})

	removeTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RemoveTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		OutputFields: graphql.Fields{
			"deletedTodoId": &graphql.Field{
				Type: graphql.ID,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoId"], nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			id, _ := inputMap["id"].(string)
			resolvedId := relay.FromGlobalID(id)
			RemoveTodo(resolvedId.ID)
			return map[string]interface{}{
				"todoId": id,
			}, nil
		},
	})
	renameTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RenameTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"todo": &graphql.Field{
				Type: todoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					return GetTodo(todoId), nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			id, _ := inputMap["id"].(string)
			resolvedId := relay.FromGlobalID(id)
			title, _ := inputMap["title"].(string)
			RenameTodo(resolvedId.ID, title)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}, nil
		},
	})
	updateTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "UpdateTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"note": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"important": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"remind_at": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"repeat_every": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"todo": &graphql.Field{
				Type: todoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					return GetTodo(todoId), nil
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetViewer(), nil
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			id, _ := inputMap["id"].(string)
			resolvedId := relay.FromGlobalID(id)
			title, _ := inputMap["title"].(string)
			note, _ := inputMap["note"].(string)
			important, _ := inputMap["important"].(string)
			remind_at, _ := inputMap["remind_at"].(string)
			repeat_every, _ := inputMap["repeat_every"].(string)
			UpdateTodo(id, title, note, important, remind_at, repeat_every)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}, nil
		},
	})
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addTodo":              addTodoMutation,
			"changeTodoStatus":     changeTodoStatusMutation,
			"markAllTodos":         markAllTodosMutation,
			"removeCompletedTodos": removeCompletedTodosMutation,
			"removeTodo":           removeTodoMutation,
			"renameTodo":           renameTodoMutation,
			"updateTodo":           updateTodoMutation,
		},
	})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
}
