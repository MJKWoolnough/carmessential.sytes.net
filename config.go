package main

import "database/sql"

var Config config

type config struct {
	data           map[string]string
	update, insert *sql.Stmt
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
	return c.data[key]
}

func (c *config) Set(key, value string) {
	DB.Lock()
	if _, ok := c.data[key]; ok {
		c.update.Exec(value, key)
	} else {
		c.insert.Exec(key, value)
	}
	DB.Unlock()
}
