package dao

import (
	"encoding/json"
	"fmt"

	bolt "github.com/coreos/bbolt"
	models "github.com/toasterlint/DAWS/common/models"
)

// DAO Data Access Object
type DAO struct {
	Database string
}

var db *bolt.DB

const (
	// COLLECTIONPEOPLE People collection to use in DB
	COLLECTIONPEOPLE = "people"
	// COLLECTIONBUILDING Building collection to use in DB
	COLLECTIONBUILDING = "building"
	// COLLECTIONCITY Building collection to use in DB
	COLLECTIONCITY = "city"
)

//Open opens DB
func (b *DAO) Open() error {
	var err error
	db, err = bolt.Open(b.Database, 0600, nil)
	if err != nil {
		return err
	}
	return nil
}

//Close closes DB
func (b *DAO) Close() {
	db.Close()
}

//CreateCity Create's a city within the DB
func (b *DAO) CreateCity(city models.City) error {
	err := db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(COLLECTIONCITY))
		if err != nil {
			return fmt.Errorf("create bucket failed: %s", err)
		}
		cityJSON, _ := json.Marshal(city)
		err = b.Put([]byte(city.ID), []byte(cityJSON))
		if err != nil {
			return fmt.Errorf("unable to insert city: %s", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
