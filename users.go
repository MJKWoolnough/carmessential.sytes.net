package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"math/rand"

	"github.com/MJKWoolnough/errors"
)

const saltLength = 16

var users Users

type Users struct {
	userID, userHash, createUser, updatePassword, updateEmail, getUserName *sql.Stmt
}

func (u *Users) init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [User]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [EmailAddress] TEXT NOT NULL, [Password] BLOB NOT NULL, [Phone] TEXT NOT NULL DEFAULT '');")
	if err != nil {
		return err
	}
	for _, stmt := range [...]struct {
		Stmt  **sql.Stmt
		Query string
	}{
		{&u.userID, "SELECT [ID] FROM [User] WHERE [EmailAddress] = ?;"},
		{&u.userHash, "SELECT [Password] FROM [User] WHERE [ID] = ?;"},
		{&u.createUser, "INSERT INTO [User]([Name], [EmailAddress], [Password]) VALUES (?, ?, ?);"},
		{&u.updatePassword, "UPDATE [User] SET [Password] = ? WHERE [ID] = ?;"},
		{&u.updateEmail, "UPDATE [User] SET [EmailAddress] = ? WHERE [ID] = ?;"},
		{&u.getUserName, "SELECT [Name] FROM [User] WHERE [ID] = ?;"},
	} {
		*stmt.Stmt, err = db.Prepare(stmt.Query)
		if err != nil {
			return err
		}
	}
	return nil
}

func passwordHash(salt []byte, password string) []byte {
	hash := sha256.Sum256(append(salt, password...))
	return append(salt, hash[:]...)
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
	DB.Lock()
	err := u.userID.QueryRow(emailAddress).Scan(&id)
	DB.Unlock()
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

func comparePassword(password string, saltedHash []byte) error {
	pass := passwordBuffer(password)
	copy(pass, saltedHash)
	if !bytes.Equal(saltedHash, passwordHash(pass, password)) {
		return ErrInvalidPassword
	}
	return nil
}

func (u *Users) CreateUser(name, emailAddress, password, phone string) (int64, error) {
	salt := sql.RawBytes(passwordBuffer(password))
	for n := range salt {
		salt[n] = byte(rand.Intn(256))
	}
	saltedHash := passwordHash(salt, password)

	var id int64
	DB.Lock()
	res, err := u.createUser.Exec(name, emailAddress, saltedHash, phone)
	if err == nil {
		id, err = res.LastInsertId()
	}
	DB.Unlock()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *Users) LoginUser(id int64, password string) error {
	DB.Lock()
	hash, err := u.UserHash(id)
	if err == nil {
		err = comparePassword(password, hash)
	}
	DB.Unlock()
	return err
}

func (u *Users) UpdateUserPassword(id int64, oldPassword, newPassword string) error {
	DB.Lock()
	saltedHash, err := u.UserHash(id)
	if err == nil {
		err = comparePassword(oldPassword, saltedHash)
		if err == nil {
			saltedHash = passwordHash(saltedHash[:saltLength], newPassword)
			_, err = u.updatePassword.Exec(saltedHash, id)
		}
	}
	DB.Unlock()
	return err
}

func (u *Users) UpdateUserEmail(id int64, emailAddress, password string) error {
	DB.Lock()
	saltedHash, err := u.UserHash(id)
	if err == nil {
		err = comparePassword(password, saltedHash)
		if err == nil {
			_, err = u.updateEmail.Exec(emailAddress, &id)
		}
	}
	DB.Unlock()
	return err
}

func (u *Users) GetUserName(id int) (string, error) {
	var username string
	DB.Lock()
	err := u.getUserName.QueryRow(id).Scan(&username)
	DB.Unlock()
	if err != nil {
		return "", err
	}
	return username, nil
}

// Errors
var (
	ErrInvalidPassword = errors.Error("invalid password")
)
