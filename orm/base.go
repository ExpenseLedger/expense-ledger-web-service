package orm

import (
	"log"
	"reflect"
	"sync"

	"github.com/expenseledger/web-service/db"
)

// BaseMapper ...
type BaseMapper struct {
	once       sync.Once
	modelType  reflect.Type
	insertStmt string
	deleteStmt string
	oneStmt    string
	updateStmt string
	manyStmt   string
	clearStmt  string
}

// Insert ...
func (mapper *BaseMapper) Insert(obj interface{}) (interface{}, error) {
	return worker(obj, mapper.modelType, mapper.insertStmt, "Error inserting")
}

// Delete ...
func (mapper *BaseMapper) Delete(obj interface{}) (interface{}, error) {
	return worker(obj, mapper.modelType, mapper.deleteStmt, "Error deleting")
}

// One ...
func (mapper *BaseMapper) One(obj interface{}) (interface{}, error) {
	return worker(obj, mapper.modelType, mapper.oneStmt, "Error geting")
}

func (mapper *BaseMapper) Update(obj interface{}) (interface{}, error) {
	return worker(obj, mapper.modelType, mapper.updateStmt, "Error geting")
}

// Many ...
func (mapper *BaseMapper) Many(obj interface{}) (interface{}, error) {
	return sliceWorker(
		obj,
		mapper.modelType,
		mapper.manyStmt,
		"Error selecting",
	)
}

// Clear ...
func (mapper *BaseMapper) Clear() (interface{}, error) {
	return sliceWorker(
		struct{}{},
		mapper.modelType,
		mapper.clearStmt,
		"Error clearing",
	)
}

func worker(
	obj interface{},
	t reflect.Type,
	sqlStmt string,
	logMsg string,
) (interface{}, error) {
	stmt, err := db.Conn().PrepareNamed(sqlStmt)
	if err != nil {
		log.Println(logMsg, err)
		return nil, err
	}

	newObj := reflect.New(t).Interface()
	if err := stmt.Get(newObj, obj); err != nil {
		log.Println(logMsg, err)
		return nil, err
	}
	return newObj, nil
}

func sliceWorker(
	obj interface{},
	t reflect.Type,
	sqlStmt string,
	logMsg string,
) (interface{}, error) {
	stmt, err := db.Conn().PrepareNamed(sqlStmt)
	if err != nil {
		log.Println(logMsg, err)
		return nil, err
	}

	sliceType := reflect.SliceOf(t)
	newSlice := reflect.MakeSlice(sliceType, 0, 0)
	resultSet := reflect.New(newSlice.Type()).Interface()
	if err := stmt.Select(resultSet, obj); err != nil {
		log.Println(logMsg, err)
		return nil, err
	}
	return resultSet, nil
}
