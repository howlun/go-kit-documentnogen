package models

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DBClient interface {
	CurrentDB() *mgo.Database
	DialWithInfo() error
	Close()
}

type mongoDBClient struct {
	DBSession    *mgo.Session
	DBUrl        string
	DBName       string
	AuthUsername string
	AuthPassword string
}

func NewDBClient(dbUrl string, dbName string, authUsername string, authPassword string) (db DBClient) {
	// TODO: Implement initialization of service
	db = &mongoDBClient{
		DBUrl:        dbUrl,
		DBName:       dbName,
		AuthUsername: authUsername,
		AuthPassword: authPassword,
	}
	return db
}

func (m *mongoDBClient) CurrentDB() *mgo.Database {
	return m.DBSession.DB(m.DBName)
}

func (m *mongoDBClient) DialWithInfo() error {
	var err error

	if m.DBUrl == "" {
		return errors.New("DB Url is empty")
	}

	if m.DBName == "" {
		return errors.New("DB Name is empty")
	}

	// We need this object to establish a session to our MongoDB.
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{m.DBUrl},
		Timeout:  60 * time.Second,
		Database: m.DBName,
		Username: m.AuthUsername,
		Password: m.AuthPassword,
	}

	m.DBSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDBClient) Close() {
	m.DBSession.Close()
}

type DocNo struct {
	Prefix          string `bson:"prefix"`
	Path            string `bson:"path"`
	NextSeqNo       int64  `bson:"nextseqno"`
	RecordTimestamp int64  `bson:"recordtimestamp"` // Unix timestamp
}

type DocNoRepository interface {
	GetByPath(docCode string, orgCode string, path string) (doc *DocNo, err error)
	UpdateByPath(orgCode string, doc *DocNo) (updated *DocNo, err error)
}

type docNoRepository struct {
	DB DBClient
}

func NewDocNoRepository(dbClient DBClient) (r DocNoRepository) {
	r = &docNoRepository{
		DB: dbClient,
	}
	return r
}

func (d *docNoRepository) GetByPath(docCode string, orgCode string, path string) (doc *DocNo, err error) {
	if docCode == "" {
		return nil, errors.New("Doc Code is empty")
	}

	if orgCode == "" {
		return nil, errors.New("Organization Code is empty")
	}

	if d.DB == nil {
		return nil, errors.New("DB Client is Nil")
	}

	// Dial to DB with extra info
	err = d.DB.DialWithInfo()
	if err != nil {
		return nil, fmt.Errorf("Failed to establish connection to Mongo Server: %s", err.Error())
	}
	defer d.DB.Close()

	// the document is group by collection (organization code)
	// perform find doc by colletion and paramter: Path (no unique ID here)
	collection := d.DB.CurrentDB().C(orgCode)
	if collection == nil {
		return nil, fmt.Errorf("Collection is nil with Org Code=%s", orgCode)
	}
	//fmt.Println(collection.FullName)

	err = collection.Find(bson.M{"prefix": docCode, "path": path}).One(&doc)
	// if has error and error not equal to document Not Found
	if err != nil && err.Error() != "not found" {
		return nil, fmt.Errorf("Error finding document with Path=%s Error=%s", path, err.Error())
	}
	// if no document found, we create new
	fmt.Println("if no document found, we create new")
	if doc == nil {
		// create new document and start with 1
		fmt.Println("create new document and start with 1")
		doc = &DocNo{
			Prefix:          docCode,
			Path:            path,
			NextSeqNo:       1,
			RecordTimestamp: time.Now().Unix(),
		}

		// insert the new document to collection
		fmt.Println("insert the new document to collection")
		err = collection.Insert(doc)
		if err != nil {
			return nil, fmt.Errorf("Error inserting document with Prefix=%s Path=%s Error=%s", docCode, path, err.Error())
		}

	}
	fmt.Println(doc)
	return doc, nil
}

func (d *docNoRepository) UpdateByPath(orgCode string, doc *DocNo) (updated *DocNo, err error) {
	if doc == nil {
		return nil, errors.New("Document to be updated is nil")
	}

	if orgCode == "" {
		return nil, errors.New("Organization Code is empty")
	}

	if doc.Prefix == "" {
		return nil, errors.New("Document Prefix is empty")
	}

	if doc.NextSeqNo == 0 {
		return nil, errors.New("Document Next Sequence No is empty")
	}

	// Dial to DB with extra info
	err = d.DB.DialWithInfo()
	if err != nil {
		return nil, fmt.Errorf("Failed to establish connection to Mongo Server: %s", err.Error())
	}
	defer d.DB.Close()

	// the document is group by collection (organization code)
	// perform find doc by colletion and paramter: Path (no unique ID here)
	collection := d.DB.CurrentDB().C(orgCode)
	if collection == nil {
		return nil, fmt.Errorf("Collection is nil with Org Code=%s", orgCode)
	}

	// partial update the document to collection
	fmt.Println("update the document to collection")
	err = collection.Update(bson.M{"prefix": doc.Prefix, "path": doc.Path}, bson.M{"$set": bson.M{"nextseqno": doc.NextSeqNo, "recordtimestamp": doc.RecordTimestamp}})
	if err != nil {
		return nil, fmt.Errorf("Error updating document with Prefix=%s Path=%s Error=%s", doc.Prefix, doc.Path, err.Error())
	}
	updated = doc
	fmt.Println(doc)
	return updated, nil
}
