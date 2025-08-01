package service

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"gopkg.in/mgo.v2/bson"
)

func TestShortUUID(t *testing.T) {
	uid1 := shortuuid.New()
	uid2 := shortuuid.NewWithNamespace("test")
	t.Logf("len:%d, uid:%s", len(uid1), uid1)
	t.Logf("len:%d, uid:%s", len(uid2), uid2)
}

func TestXID(t *testing.T) {
	xid1 := xid.New()
	xid2 := xid.NewWithTime(time.Now())
	t.Logf("len:%d, uid:%s", len(xid1), xid1)
	t.Logf("len:%d, uid:%s", len(xid2), xid2)
}

func TestKSUID(t *testing.T) {
	uid1 := ksuid.New()
	uid2, err := ksuid.NewRandom()
	if err != nil {
		t.Error(err)
	}
	uid3, err := ksuid.NewRandomWithTime(time.Now())
	if err != nil {
		t.Error(err)
	}
	t.Logf("len:%d, uid:%s", len(uid1), uid1)
	t.Logf("len:%d, uid:%s", len(uid2), uid2)
	t.Logf("len:%d, uid:%s", len(uid3), uid3)
}

func TestULID(t *testing.T) {
	uid1 := ulid.Make()
	reader := bytes.NewReader([]byte("test"))
	uid2, err := ulid.New(ulid.Now(), reader)
	if err != nil {
		t.Error(err)
	}
	t.Logf("len:%d, uid:%s", len(uid1), uid1)
	t.Logf("len:%d, uid:%s", len(uid2), uid2)
}

func TestUUID(t *testing.T) {
	uid := uuid.New()
	t.Log(uid.String())
}

func TestObjectID(t *testing.T) {
	id := bson.NewObjectId()
	t.Log(len(id.Hex()), id.Hex())
}

func Test(t *testing.T) {
	t.Run("shortuuid", TestShortUUID)
	t.Run("xid", TestXID)
	t.Run("ksuid", TestKSUID)
	t.Run("ulid", TestULID)
}
