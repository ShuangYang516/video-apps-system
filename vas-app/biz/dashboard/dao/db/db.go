package db

import (
	mgoutil "github.com/qiniu/db/mgoutil.v3"
	mgo "gopkg.in/mgo.v2"
)

type collections struct {
	Event mgoutil.Collection `coll:"vas_event"`
	Task mgoutil.Collection `coll:"vas_tasks"`
}

var colls *collections

func Init(cfg *mgoutil.Config, spl int) error {
	c := &collections{}
	mgoSession, e := mgoutil.Open(c, cfg)
	if e != nil {
		return e
	}

	if spl != 0 {
		mgoSession.SetPoolLimit(spl)
	}
	colls = c
	return nil
}

func IsNotFoundError(e error) bool {
	return e == mgo.ErrNotFound
}

func IsDupError(e error) bool {
	return mgo.IsDup(e)
}
