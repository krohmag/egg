package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Datastore is an interface for interacting with a database
type Datastore interface {
	Ping() bool
	Transaction(ctx context.Context) (Transaction, error)
}

// Transaction is an interface for running transactions against a datastore
type Transaction interface {
	Commit() error
	Rollback() error

	CreateOrUpdateUser(user User) (User, error)

	GetUsers() (Users, error)
	GetUsersByDiscordName(discordName string) (Users, error)
	GetUserByEggIncUserID(eggIncUserID string) (User, error)

	DeleteUser(user User) error
}

// User is the struct representation of a database table for storing user information
type User struct {
	EggIncID        string  `json:"egg_inc_id" gorm:"egg_inc_id;primarykey,unique;not null"`
	DiscordName     string  `json:"discord_name" gorm:"discord_name;not null"`
	GameAccountName string  `json:"game_account_name" gorm:"game_account_name;unique"`
	SoulFood        int32   `json:"soul_food" gorm:"soul_food"`
	ProphecyBonus   int32   `json:"prophecy_bonus" gorm:"prophecy_bonus"`
	SoulEggs        float64 `json:"soul_eggs" gorm:"soul_eggs"`
	ProphecyEggs    int32   `json:"prophecy_eggs" gorm:"prophecy_eggs"`

	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Users is a slice of the User type
type Users []User

func (u Users) GetEggIncIDs() (ids []string) {
	for _, user := range u {
		ids = append(ids, user.EggIncID)
	}
	return
}

// Database implements the Datastore interface
type Database struct {
	DB *gorm.DB
}

// Ping ensures database connectivity
func (d Database) Ping() bool {
	var result int
	d.DB.Raw("SELECT 1+1").Scan(&result)

	if result != 2 {
		return false
	}

	return true
}

// Transaction returns an implementation of the Transaction interface
func (d Database) Transaction(ctx context.Context) (Transaction, error) {
	tx := d.DB.WithContext(ctx)
	if tx.Error != nil {
		return nil, tx.Error
	}

	result := tx.Begin()
	if result.Error != nil {
		return nil, result.Error
	}

	return Txn{Client: result}, nil
}

// Txn implements the Transaction interface
type Txn struct {
	Client *gorm.DB
}

// Commit commits a transaction
func (t Txn) Commit() error {
	return t.Client.Commit().Error
}

// Rollback rolls back a transaction
func (t Txn) Rollback() error {
	return t.Client.Rollback().Error
}

// CreateOrUpdateUser adds or updates a user to the datastore
func (t Txn) CreateOrUpdateUser(user User) (User, error) {
	var userTemplate User
	switch err := t.Client.Where("egg_inc_id = ?", user.EggIncID).First(&userTemplate).Error; {
	case err == nil:
		// update because the record already exists
		if err = t.Client.Where("egg_inc_id = ?", user.EggIncID).Updates(&user).Error; err != nil {
			return user, err
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// create because the record does not exist
		if err = t.Client.Create(&user).Error; err != nil {
			return user, err
		}
	default:
		// some other unexpected error
		return user, err
	}

	err := t.Client.Where("egg_inc_id = ?", user.EggIncID).First(&userTemplate).Error
	return userTemplate, err
}

// GetUsers returns all user info
func (t Txn) GetUsers() (Users, error) {
	var users Users
	if err := t.Client.Find(&users).Error; err != nil {
		return Users{}, err
	}

	return users, nil
}

// GetUsersByDiscordName returns all user for a given discord username
func (t Txn) GetUsersByDiscordName(discordName string) (Users, error) {
	var users Users
	if err := t.Client.Where("discord_name = ?", discordName).Find(&users).Error; err != nil {
		return Users{}, err
	}

	if len(users) == 0 {
		return Users{}, errors.New(fmt.Sprintf("no records found for provided Discord name: %s", discordName))
	}

	return users, nil
}

// GetUserByEggIncUserID returns a user for a given Egg, Inc. user ID
func (t Txn) GetUserByEggIncUserID(eggIncUserID string) (User, error) {
	var user User
	if err := t.Client.Where("egg_inc_id = ?", eggIncUserID).First(&user).Error; err != nil {
		return User{}, err
	}

	return user, nil
}

// DeleteUser removes a user from the datastore
func (t Txn) DeleteUser(user User) error {
	return t.Client.Where("egg_inc_id = ?", user.EggIncID).Delete(&user).Error
}

// ConnectDatabase stolen from tinkerbell-cerberus and Wyatt ;-) for connecting to different DB types
func ConnectDatabase(url string, silenceTransactionLogs bool) (*gorm.DB, error) {
	var backendPath gorm.Dialector
	switch url {
	// https://www.sqlite.org/inmemorydb.html
	case "sqlite-in-memory":
		backendPath = sqlite.Open("file::memory:")
	case "sqlite-file":
		backendPath = sqlite.Open("./db.sqlite")
	default:
		backendPath = postgres.Open(url)
	}

	var config gorm.Config
	switch silenceTransactionLogs {
	case true:
		config = gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	default:
		config = gorm.Config{}
	}

	db, err := gorm.Open(backendPath, &config)
	if err != nil {
		return nil, err
	}

	return db, nil
}
