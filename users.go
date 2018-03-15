package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/errors"
)

const saltLength = 16

var Users users

type users struct {
	userID, userHash, createUser, updatePassword, updateEmail, getUserName *sql.Stmt
}

func (u *users) init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [User]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [EmailAddress] TEXT NOT NULL, [Password] BLOB NOT NULL, [Phone] TEXT NOT NULL DEFAULT '');")
	if err != nil {
		return errors.WithContext("error creating User table: ", err)
	}
	for _, stmt := range [...]struct {
		Stmt  **sql.Stmt
		Query string
	}{
		{&u.userID, "SELECT [ID] FROM [User] WHERE [EmailAddress] = ?;"},
		{&u.userHash, "SELECT [Password] FROM [User] WHERE [ID] = ?;"},
		{&u.createUser, "INSERT INTO [User]([Name], [EmailAddress], [Password], [Phone]) VALUES (?, ?, ?, ?);"},
		{&u.updatePassword, "UPDATE [User] SET [Password] = ? WHERE [ID] = ?;"},
		{&u.updateEmail, "UPDATE [User] SET [EmailAddress] = ? WHERE [ID] = ?;"},
		{&u.getUserName, "SELECT [Name] FROM [User] WHERE [ID] = ?;"},
	} {
		*stmt.Stmt, err = db.Prepare(stmt.Query)
		if err != nil {
			return errors.WithContext(fmt.Sprintf("error preparing User statement %q: ", stmt.Query), err)
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

func (u *users) UserID(emailAddress string) (int64, error) {
	var id int64
	DB.Lock()
	err := u.userID.QueryRow(emailAddress).Scan(&id)
	DB.Unlock()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *users) UserHash(id int64) ([]byte, error) {
	passHash := make([]byte, saltLength, saltLength+sha256.Size)
	err := u.userHash.QueryRow(id).Scan(&passHash)
	return passHash, err
}

func comparePassword(password string, saltedHash []byte) error {
	pass := passwordBuffer(password)
	copy(pass, saltedHash[:saltLength])
	if !bytes.Equal(saltedHash, passwordHash(pass[:saltLength], password)) {
		return ErrInvalidPassword
	}
	return nil
}

func (u *users) CreateUser(name, emailAddress, password, phone string) (int64, error) {
	salt := passwordBuffer(password)
	for n := range salt[:saltLength] {
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

func (u *users) LoginUser(id int64, password string) error {
	DB.Lock()
	hash, err := u.UserHash(id)
	if err == nil {
		err = comparePassword(password, hash)
	}
	DB.Unlock()
	return err
}

func (u *users) UpdateUserPassword(id int64, oldPassword, newPassword string) error {
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

func (u *users) UpdateUserEmail(id int64, emailAddress, password string) error {
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

func (u *users) GetUserName(id int64) (string, error) {
	var username string
	DB.Lock()
	err := u.getUserName.QueryRow(id).Scan(&username)
	DB.Unlock()
	if err != nil {
		return "", err
	}
	return username, nil
}

func (u *users) IsAdmin(id int64) bool {
	idstr := strconv.Itoa(int(id))
	for _, ids := range strings.Split(Config.Get("admins"), " ") {
		if ids == idstr {
			return true
		}
	}
	return false
}

// Errors
var (
	ErrInvalidPassword = errors.Error("invalid password")
)
