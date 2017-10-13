package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"math/rand"
)

const saltLength = 16

var users Users

type Users struct {
	*database
	userID, userHash, createUser, updatePassword, updateEmail, getUserName *sql.Stmt
}

func (u *Users) init(db *database) error {
	u.database = db
	_, err := db.db.Exec("CREATE TABLE IF NOT EXIST [User]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [EmailAddress] TEXT NOT NULL, [Password] BLOB NOT NULL);")
	if err != nil {
		return err
	}
	for stmt, query := range map[**sql.Stmt]string{
		&u.userID:         "SELECT [ID] FROM [User] WHERE [EmailAddress] = ?;",
		&u.userHash:       "SELECT [Password] FROM [User] WHERE [ID] = ?;",
		&u.createUser:     "INSERT INTO [User]([Name], [EmailAddress], [Password]) VALUES (?, ?, ?);",
		&u.updatePassword: "UPDATE [User] SET [Password] = ? WHERE [ID] = ?;",
		&u.updateEmail:    "UPDATE [User] SET [EmailAddress] = ? WHERE [ID] = ?;",
		&getUserName:      "SELECT [Name] FROM [User] WHERE [ID] = ?;",
	} {
		*stmt, err = db.db.Prepare(query)
		if err != nil {
			return err
		}
	}
	u.mu = &db.mu
	return nil
}

func passwordHash(salt []byte, password string) []byte {
	hash := sha256.Sum256(append(salt, password...))
	return append(salt, hash[:])
}

func passwordBuffer(password string) []byte {
	l := len(password)
	if l < sha256.Size {
		l = sha256.Size
	}
	return make([]byte, saltLength, saltLength+l)
}

func (u *Users) UserID(emailAddress string) (int64, error) {
	var id int64
	u.mu.Lock()
	err := u.userID.QueryRow(emailAddress).Scan(&id)
	u.mu.Unlock()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *Users) UserHash(id int64) (sql.RawBytes, error) {
	passHash := make(sql.RawBytes, saltLength, saltLength+sha256.Size)
	err := u.userHash.QueryRow(id).Scan(&passHash)
	return passHash, err
}

func comparePassword(id int64, password string, saltedHash []byte) error {
	pass := passwordBuffer(password)
	copy(pass, saltedHash)
	if !bytes.Compare(saltedHash, passwordHash(append(pass, password...))) {
		return ErrInvalidPassword
	}
	return nil
}

func (u *Users) CreateUser(name, emailAddress, password string) (int64, error) {
	salt := sql.RawBytes(passwordBuffer(password))
	for n := range salt {
		salt[n] = byte(rand.Intn(256))
	}
	saltedHash := passwordHash(salt, password)

	var id int64
	u.mu.Lock()
	res, err := u.createUser.Exec(name, emailAddress, saltedHash)
	if err == nil {
		id, err = res.LastInsertId()
	}
	u.mu.Unlock()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *Users) LoginUser(id int64, password string) error {
	u.mu.Lock()
	err := comparePassword(id, password, u.userHash(id))
	u.mu.Unlock()
	return err
}

func (u *Users) UpdateUserPassword(id int64, oldPassword, newPassword string) error {
	u.mu.Lock()
	saltedHash := d.userHash(id)
	err := d.comparePassword(id, oldPassword, saltedHash)
	if err == nil {
		saltedHash = passwordHash(saltedHash[:saltLength], newPassword)
		err = u.updatePassword.Exec(saltedHash, id)
	}
	u.mu.Unlock()
	return err
}

func (d *database) UpdateUserEmail(id int64, emailAddress, password string) error {
	u.mu.Lock()
	err := comparePassword(id, password, u.userHash(id))
	if err == nil {
		err = u.updateEmail.Exec(emailAddress, &id)
	}
	u.mu.Unlock()
	return err
}

func (d *database) GetUserName(id int) (string, error) {
	var username string
	u.mu.Lock()
	err := u.getUserName.QueryRow(id).Scan(&username)
	u.mu.Unlock()
	if err != nil {
		return "", err
	}
	return username, nil
}
