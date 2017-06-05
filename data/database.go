package data

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const DB = "todos"
const TODO_COLLECTION = "todo"
const (
	ID           = "_id"
	COMPLETE     = "complete"
	DELETE       = "delete"
	CREATED_BY   = "created_by"
	TITLE        = "title"
	NOTE         = "note"
	CREATED_AT   = "created_at"
	UPDATED_AT   = "updated_at"
	REMIND_AT    = "remind_at"
	REPEAT_EVERY = "repeat_every"
	IMPORTANT    = "important"
)

var (
	mongoSession *mgo.Session
	err          error
)

func init() {
	mongoSession, err = ConnnectMongo()
	if err != nil {
		panic(err)
	}
}

func ConnnectMongo() (*mgo.Session, error) {
	mongoSession, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Timeout:  1800 * time.Second,
		Database: "admin",
		Username: "",
		Password: "",
	})
	if err != nil {
		return &mgo.Session{}, err
	}
	mongoSession.SetMode(mgo.Monotonic, true)
	return mongoSession, nil

}

// Mock authenticated ID
const ViewerId = "me"

// Model structs
type Todo struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Title       string        `json:"title" bson:"title"`
	Note        string        `json:"note" bson:"note"`
	Complete    bool          `json:"complete" bson:"complete"`
	Delete      bool          `json:"delete" bson:"delete"`
	CreatedBy   string        `json:"created_by" bson:"created_by"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
	RemindAt    time.Time     `json:"remind_at" bson:"remind_at,omitempty"`
	RepeatEvery string        `json:"repeat_every" bson:"repeat_every"`
	Important   string        `json:"important" bson:"important"`
}

type User struct {
	ID string `json:"id"`
}

// Mock data
var viewer = &User{ViewerId}
var usersById = map[string]*User{
	ViewerId: viewer,
}
var todosById = map[string]*Todo{}
var todoIdsByUser = map[string][]string{
	ViewerId: []string{},
}
var nextTodoId = 0

// Data methods

func AddTodo(title string, important string, complete bool) string {
	created_at := time.Now()
	todo := &Todo{
		ID:        bson.NewObjectId(),
		Title:     title,
		Important: important,
		CreatedAt: created_at,
		UpdatedAt: created_at,
		Complete:  complete,
		Delete:    false,
		CreatedBy: ViewerId,
	}
	session := mongoSession.Clone()
	defer session.Close()
	session.DB(DB).C(TODO_COLLECTION).Upsert(bson.M{ID: todo.ID}, todo)
	return todo.ID.Hex()
}

func GetTodo(id string) *Todo {
	fmt.Println("GetTodo", id)
	// if todo, ok := todosById[id]; ok {
	// 	return todo
	// }
	if bson.IsObjectIdHex(id) {
		session := mongoSession.Clone()
		defer session.Close()
		todo := &Todo{}
		session.DB(DB).C(TODO_COLLECTION).Find(bson.M{ID: bson.ObjectIdHex(id), DELETE: false}).One(todo)
		return todo
	}
	return nil
}

func GetTodos(status string) []*Todo {
	todos := []*Todo{}
	session := mongoSession.Clone()
	defer session.Close()
	var query bson.M
	switch status {
	case "completed":
		query = bson.M{CREATED_BY: ViewerId, COMPLETE: true, DELETE: false}
	case "incomplete":
		query = bson.M{CREATED_BY: ViewerId, COMPLETE: false, DELETE: false}
	case "any":
		fallthrough
	default:
		query = bson.M{CREATED_BY: ViewerId, DELETE: false}
	}
	err := session.DB(DB).C(TODO_COLLECTION).Find(query).Sort(fmt.Sprintf("-%s", UPDATED_AT)).All(&todos)
	fmt.Println("GetTodos", query, err)
	for _, v := range todos {
		fmt.Println(*v)
	}
	return todos
}

func GetUser(id string) *User {
	return &User{ID: ViewerId}
}

func GetViewer() *User {
	return GetUser(ViewerId)
}

func ChangeTodoStatus(id string, complete bool) {
	fmt.Println("ChangeTodoStatus", id)
	if bson.IsObjectIdHex(id) {
		session := mongoSession.Clone()
		defer session.Close()
		session.DB(DB).C(TODO_COLLECTION).UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{COMPLETE: complete}})
	}

}

func MarkAllTodos(complete bool) []string {
	changedTodoIds := []string{}
	todosId_ := []bson.ObjectId{}
	session := mongoSession.Clone()
	defer session.Close()
	todos := GetTodos("any")
	for _, v := range todos {
		todosId_ = append(todosId_, (*v).ID)
		changedTodoIds = append(changedTodoIds, (*v).ID.Hex())
	}
	session.DB(DB).C(TODO_COLLECTION).UpdateAll(bson.M{ID: bson.M{"$in": todosId_}}, bson.M{"$set": bson.M{COMPLETE: complete}})
	return changedTodoIds
}

func RemoveTodo(id string) {
	fmt.Println("RemoveTodo", id)
	if bson.IsObjectIdHex(id) {
		session := mongoSession.Clone()
		defer session.Close()
		session.DB(DB).C(TODO_COLLECTION).UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{DELETE: true}})
	}

}

func RemoveCompletedTodos() []string {
	todosIdRemoved := []string{}
	todosId_ := []bson.ObjectId{}
	session := mongoSession.Clone()
	defer session.Close()
	todos := GetTodos("completed")
	for _, v := range todos {
		todosId_ = append(todosId_, (*v).ID)
		todosIdRemoved = append(todosIdRemoved, (*v).ID.Hex())
	}
	session.DB(DB).C(TODO_COLLECTION).UpdateAll(bson.M{ID: bson.M{"$in": todosId_}}, bson.M{"$set": bson.M{DELETE: true}})
	return todosIdRemoved
}

func RenameTodo(id string, title string) {
	fmt.Println("RenameTodo", id)
	if bson.IsObjectIdHex(id) {
		session := mongoSession.Clone()
		defer session.Close()
		session.DB(DB).C(TODO_COLLECTION).UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{TITLE: title, UPDATED_AT: time.Now()}})
	}
}
func UpdateTodo(id string, title string, note string, important string, remind_at string, reapeat_every string) {
	fmt.Println("RenameTodo", id)
	if bson.IsObjectIdHex(id) {
		session := mongoSession.Clone()
		defer session.Close()
		var remind_at_time time.Time
		remind_at_time_tmp, err := time.Parse(time.RFC3339, remind_at)
		if err != nil {
			remind_at_time = time.Now().AddDate(0, 0, 1)
		} else {
			remind_at_time = remind_at_time_tmp
		}
		session.DB(DB).C(TODO_COLLECTION).UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{TITLE: title, NOTE: note, IMPORTANT: important, remind_at: remind_at_time, REPEAT_EVERY: reapeat_every, UPDATED_AT: time.Now()}})
	}
}

func TodosToSliceInterface(todos []*Todo) []interface{} {
	todosIFace := []interface{}{}
	for _, todo := range todos {
		todosIFace = append(todosIFace, todo)
	}
	return todosIFace
}
