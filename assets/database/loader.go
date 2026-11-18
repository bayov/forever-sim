package database

import (
	_ "embed"

	"github.com/wowsims/classic/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

//go:embed db.bin
var dbBytes []byte

func Load() *proto.UIDatabase {
	db := &proto.UIDatabase{}
	if err := googleProto.Unmarshal(dbBytes, db); err != nil {
		panic(err)
	}
	return db
}

// LoadSim decodes only the part of the database the sim reads (see SimDatabaseFile).
func LoadSim() *proto.SimDatabaseFile {
	db := &proto.SimDatabaseFile{}
	if err := (googleProto.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(dbBytes, db); err != nil {
		panic(err)
	}
	return db
}
