package database

type Database interface {
	get(string) interface{}
	put(string, interface{}) int
}

func New(dbType string) Database {
	if dbType == "file" {
		return nil
	}
	if dbType == "redis" {
		return nil
	}
	return nil
}
