package cloudprint

import (
	"errors"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TokenStore interface {
	Get(clientID string) (token string, err error)
	Set(clientID, token string) (err error)
}

type MemoryStore struct {
	sm sync.Map
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Get(clientID string) (token string, err error) {
	value, ok := s.sm.Load(clientID)
	if ok {
		token = value.(string)
	}
	return
}

func (s *MemoryStore) Set(clientID, token string) (err error) {
	s.sm.Store(clientID, token)
	return
}

type GormModel struct {
	ClientID string `gorm:"primaryKey"`
	Token    string
}

type GormStore struct {
	db        *gorm.DB
	tableName string
}

func NewGormStore(db *gorm.DB, tableName string) *GormStore {
	return &GormStore{
		db:        db,
		tableName: tableName,
	}
}

func (s *GormStore) Get(clientID string) (token string, err error) {
	model := &GormModel{}
	err = s.db.Table(s.tableName).Where("client_id = ?", clientID).First(model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	token = model.Token
	return
}

func (s *GormStore) Set(clientID, token string) (err error) {
	model := &GormModel{
		ClientID: clientID,
		Token:    token,
	}
	err = s.db.Table(s.tableName).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "client_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"token"}),
	}).Create(model).Error
	return
}
