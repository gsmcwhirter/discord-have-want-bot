package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"

	bolt "github.com/coreos/bbolt"
	"github.com/gsmcwhirter/discord-bot-lib/snowflake"
	"github.com/gsmcwhirter/go-util/deferutil"
	"github.com/pkg/errors"
)

type config struct {
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	AllUsers bool   `mapstructure:"all_users"`
}

func start(c config) error {
	fmt.Printf("%+v\n", c)

	deps, err := createDependencies(c)
	if err != nil {
		return err
	}
	defer deps.Close()

	uid, err := snowflake.FromString(c.User)
	if err != nil {
		return errors.Wrap(err, "could not parse user id")
	}

	if c.AllUsers {
		return dumpAllUsers(deps)
	}

	if err = dumpUserCharacters(deps, uid); err != nil {
		return err
	}

	return nil
}

func dumpAllUsers(deps *dependencies) error {
	t, err := deps.db.Begin(false)
	if err != nil {
		return err
	}
	defer t.Rollback()

	fixups := map[string][]byte{}

	err = t.ForEach(func(name []byte, b *bolt.Bucket) error {
		fmt.Println(string(name))
		err := b.ForEach(func(k, v []byte) error {
			fmt.Printf("  %s\n", string(k))
			if cmdhandler.IsUserMention(string(k)) {
				fixups[string(k)[3:len(string(k))-1]] = v
			}
			return nil
		})
		fmt.Println()
		return err
	})

	if err != nil {
		return err
	}

	t.Rollback()
	if len(fixups) == 0 {
		return nil
	}

	t, err = deps.db.Begin(true)
	if err != nil {
		return err
	}
	defer t.Rollback()

	b := t.Bucket([]byte("UserRecords"))

	for k, v := range fixups {
		pU := &storage.ProtoUser{}
		proto.Unmarshal(v, pU)
		pU.Name = k
		v, err = proto.Marshal(pU)
		if err != nil {
			return err
		}

		fmt.Println(pU.Name)

		err := b.Put([]byte(k), v)
		if err != nil {
			return err
		}
		err = b.Delete([]byte(fmt.Sprintf("<@!%s>", k)))
		if err != nil {
			return err
		}
		fmt.Println(k)
	}

	t.Commit()

	return nil
}

func dumpUserCharacters(deps *dependencies, uid snowflake.Snowflake) error {
	t, err := deps.UserAPI().NewTransaction(false)
	if err != nil {
		return errors.Wrap(err, "could not get transaction")
	}
	defer deferutil.CheckDefer(t.Rollback)

	u, err := t.GetUser(uid.ToString())
	if err != nil {
		return errors.Wrap(err, "could not get user")
	}

	for _, c := range u.GetCharacters() {
		fmt.Printf("%+v\n", c)
	}
	return nil
}
