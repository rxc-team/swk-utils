package database

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MongoInsert 插入数据
func MongoInsert(collection string, data ...interface{}) error {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	if err := c.Insert(data...); err != nil {
		log.Errorf("error MongoInsert %v collection: %v", err, collection)
		return err
	}

	log.Infof("MongoInsert took: %v collection: %v len(%v)", time.Since(s), collection, len(data))
	return nil
}

// MongoUpdate 更新数据
func MongoUpdate(collection string, selector interface{}, update interface{}) error {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	if err := c.Update(selector, update); err != nil {
		log.Errorf("error MongoUpdate %v collection: %v", err, collection)
		return err
	}

	log.Infof("MongoUpdate took: %v collection: %v", time.Since(s), collection)
	return nil
}

// MongoRemove 删除集合
func MongoRemove(collection string, selector interface{}) error {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	if err := c.Remove(selector); err != nil {
		log.Errorf("error MongoRemove on selector %v", err)
		return err
	}

	log.Infof("MongoRemove took: %v", time.Since(s))
	return nil
}

// MongoCreateCollection 创建集合
func MongoCreateCollection(collection string, info *mgo.CollectionInfo) error {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	// info := mgo.CollectionInfo{ForceIdIndex: false, DisableIdIndex: true}
	if err := c.Create(info); err != nil {
		log.Errorf("error MongoCreateCollection %v", err)
		return err
	}

	log.Infof("createDatastoreCollection took: %v collectionName=%v", time.Since(s), collection)
	return nil

}

// CountCollection 从集合中获取合计数
func CountCollection(collection string, query bson.M) int {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	count, err := c.Find(query).Count()
	if err != nil {
		log.Errorf("error CountCollection %v", err)
		return 0
	}

	log.Infof("CountCollection took: %v items(%v)", time.Since(s), count)
	return count
}

// MongoCreateIndex 创建索引
func MongoCreateIndex(collection string, index mgo.Index) error {
	s, sc := BeginMongo()
	c := sc.DB(Db).C(collection)
	defer sc.Close()

	if err := c.EnsureIndex(index); err != nil {
		log.Errorf("error MongoEnsureIndex %v", err)
		return err
	}

	log.Infof("MongoEnsureIndex took: %v", time.Since(s))
	return nil
}
