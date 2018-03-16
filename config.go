package main

import (
	"database/sql"
	"sort"
	"sync"

	"github.com/MJKWoolnough/errors"
)

var Config config

type config struct {
	update, insert, remove *sql.Stmt

	lock sync.RWMutex
	data map[string]string
}

func (c *config) init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [Config]([Key] TEXT NOT NULL, [Value] TEXT NOT NULL DEFAULT '');")
	if err != nil {
		return errors.WithContext("error creating Config table: ", err)
	}
	rows, err := db.Query("SELECT [Key], [Value] FROM [Config];")
	if err != nil {
		return errors.WithContext("error get Config data: ", err)
	}
	c.data = make(map[string]string)
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			return errors.WithContext("error getting Config row data: ", err)
		}
		c.data[key] = value
	}
	c.update, err = db.Prepare("UPDATE [Config] SET [Value] = ? WHERE [Key] = ?;")
	if err != nil {
		return errors.WithContext("error preparing Config Update statement: ", err)
	}
	c.insert, err = db.Prepare("INSERT INTO [Config] ([Key], [Value]) VALUES (?, ?);")
	if err != nil {
		return errors.WithContext("error preparing Config Insert statement: ", err)
	}
	c.remove, err = db.Prepare("DELETE FROM [Config] WHERE [Key] = ?;")
	if err != nil {
		return errors.WithContext("error preparing Config Remove statement: ", err)
	}
	return nil
}

func (c *config) Get(key string) string {
	c.lock.RLock()
	value := c.data[key]
	c.lock.RUnlock()
	return value
}

func (c *config) Set(key, value string) {
	c.lock.Lock()
	DB.Lock()
	if _, ok := c.data[key]; ok {
		c.update.Exec(value, key)
	} else {
		c.insert.Exec(key, value)
	}
	DB.Unlock()
	c.data[key] = value
	c.lock.Unlock()
}

type KeyValues []KeyValue

func (k KeyValues) Len() int {
	return len(k)
}

func (k KeyValues) Less(i, j int) bool {
	return k[i].Key < k[j].Key
}

func (k KeyValues) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

type KeyValue struct {
	Key, Value string
}

func (c *config) AsSlice() KeyValues {
	Config.lock.RLock()
	vars := make(KeyValues, 0, len(Config.data))
	for key, value := range Config.data {
		vars = append(vars, KeyValue{key, value})
	}
	Config.lock.RUnlock()
	sort.Sort(vars)
	return vars
}

func (c *config) Remove(key string) {
	c.lock.Lock()
	delete(c.data, key)
	c.remove.Exec(key)
	c.lock.Unlock()
}
