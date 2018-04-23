package storage

import (
	bolt "github.com/coreos/bbolt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	errorsx "github.com/pkg/errors"
)

// ErrUserNotExist TODOC
var ErrUserNotExist = errors.New("user does not exist")

type boltUserAPI struct {
	db         *bolt.DB
	bucketName []byte
}

// NewBoltUserAPI constructs a boltDB-backed UserAPI
func NewBoltUserAPI(db *bolt.DB) (UserAPI, error) {
	b := boltUserAPI{
		db:         db,
		bucketName: []byte("UserRecords"),
	}

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(b.bucketName)
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b boltUserAPI) NewTransaction(writable bool) (UserAPITx, error) {
	tx, err := b.db.Begin(writable)
	if err != nil {
		return nil, err
	}
	return boltUserAPITx{
		bucketName: b.bucketName,
		tx:         tx,
	}, nil
}

type boltUserAPITx struct {
	bucketName []byte
	tx         *bolt.Tx
}

func (b boltUserAPITx) Commit() error {
	return b.tx.Commit()
}

func (b boltUserAPITx) Rollback() error {
	return b.tx.Rollback()
}

func (b boltUserAPITx) AddUser(name string) (User, error) {
	user, err := b.GetUser(name)
	if err == ErrUserNotExist {
		user = &boltUser{
			protoUser: &ProtoUser{Name: name},
		}
		err = nil
	}
	return user, err
}

func (b boltUserAPITx) SaveUser(user User) error {
	bucket := b.tx.Bucket(b.bucketName)

	serial, err := user.Serialize()
	if err != nil {
		return err
	}

	return bucket.Put([]byte(user.GetName()), serial)
}

func (b boltUserAPITx) GetUser(name string) (User, error) {
	bucket := b.tx.Bucket(b.bucketName)

	val := bucket.Get([]byte(name))

	if val == nil {
		return nil, ErrUserNotExist
	}

	protoUser := ProtoUser{}
	err := proto.Unmarshal(val, &protoUser)
	if err != nil {
		return nil, errorsx.Wrap(err, "user record is corrupt")
	}

	return &boltUser{&protoUser}, nil
}

func (b boltUserAPITx) GetUsers() []User {
	return nil
}
