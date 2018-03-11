package main

import (
	"database/sql"
	"sync"
)

var Config config

type config struct {
	update, insert *sql.Stmt

	lock sync.RWMutex
	data map[string]string
}

func (c *config) init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [Config]([Key] TEXT NOT NULL, [Value] TEXT NOT NULL DEFAULT '');")
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT [Key], [Value] FROM [Config];")
	if err != nil {
		return err
	}
	c.data = make(map[string]string)
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			return err
		}
		c.data[key] = value
	}
	c.update, err = db.Prepare("UPDATE [Config] SET [Value] = ? WHERE [Key] = ?;")
	if err != nil {
		return err
	}
	c.insert, err = db.Prepare("INSERT INTO [Config] ([Key], [Value]) VALUES (?, ?);")
	return err
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
