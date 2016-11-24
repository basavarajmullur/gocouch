// Tideland Go CouchDB Client - CouchDB - Unit Tests
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package couchdb_test

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/etc"
	"github.com/tideland/golib/identifier"

	"github.com/tideland/gocouch/couchdb"
)

//--------------------
// CONSTANTS
//--------------------

const (
	LargerCfg      = "{etc {couchdb {hostname localhost}{port 5984}{database tgocouch-testing-temporary}}}"
	EmptyCfg       = "{etc}"
	LocalhostCfg   = "{etc {hostname localhost}{port 5984}}"
	TestingDBCfg   = "{etc {hostname localhost}{port 5984}{database tgocouch-testing}}"
	TemporaryDBCfg = "{etc {hostname localhost}{port 5984}{database tgocouch-testing-temporary}}"
	TemplateDBcfg  = "{etc {hostname localhost}{port 5984}{database tgocouch-testing-<<DATABASE>>}}"
)

//--------------------
// TESTS
//--------------------

// TestNoConfig tests opening the database without a configuration.
func TestNoConfig(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	cdb, err := couchdb.Open(nil)
	assert.ErrorMatch(err, ".* cannot open database without configuration")
	assert.Nil(cdb)
}

// TestAllDatabases tests the retrieving of all databases.
func TestAllDatabases(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	cfg, err := etc.ReadString(LargerCfg)
	assert.Nil(err)

	cdb, err := couchdb.OpenPath(cfg, "couchdb")
	assert.Nil(err)
	ids, err := cdb.AllDatabases()
	assert.Nil(err)
	assert.True(len(ids) != 0)

	cfg, err = etc.ReadString(EmptyCfg)
	assert.Nil(err)

	cdb, err = couchdb.OpenPath(cfg, "")
	assert.Nil(err)
	ids, err = cdb.AllDatabases()
	assert.Nil(err)
	assert.True(len(ids) != 0)
}

// TestCreateDeleteDatabase tests the creation and deletion
// of a database.
func TestCreateDeleteDatabase(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	cfg, err := etc.ReadString(TemporaryDBCfg)
	assert.Nil(err)

	// Open and check existance.
	cdb, err := couchdb.Open(cfg)
	assert.Nil(err)
	has, err := cdb.HasDatabase()
	assert.Nil(err)
	assert.False(has)

	// Create and check existance,
	resp := cdb.CreateDatabase()
	assert.True(resp.IsOK())
	has, err = cdb.HasDatabase()
	assert.Nil(err)
	assert.True(has)

	// Delete and check existance.
	resp = cdb.DeleteDatabase()
	assert.True(resp.IsOK())
	has, err = cdb.HasDatabase()
	assert.Nil(err)
	assert.False(has)
}

// TestCreateDesignDocument tests creating new design documents.
func TestCreateDesignDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("create-design", assert)
	defer cleanup()

	// Create design document and check if it has been created.
	allDesignA, err := cdb.AllDesignDocuments()
	assert.Nil(err)

	ddocA := &couchdb.DesignDocument{
		ID: "_design/testing-a",
		Views: couchdb.DesignViews{
			"index-a": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('a') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp := cdb.CreateDesignDocument(ddocA)
	assert.True(resp.IsOK())
	ddocB := &couchdb.DesignDocument{
		ID: "testing-b",
		Views: couchdb.DesignViews{
			"index-b": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('b') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp = cdb.CreateDesignDocument(ddocB)
	assert.True(resp.IsOK())

	allDesignB, err := cdb.AllDesignDocuments()
	assert.Nil(err)
	assert.Equal(len(allDesignB), len(allDesignA)+2)
}

// TestReadDesignDocument tests reading design documents.
func TestReadDesignDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("read-design", assert)
	defer cleanup()

	// Create design document and read it again.
	ddocA := &couchdb.DesignDocument{
		ID: "_design/testing-a",
		Views: couchdb.DesignViews{
			"index-a": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('a') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp := cdb.CreateDesignDocument(ddocA)
	assert.True(resp.IsOK())

	ddocB, err := cdb.ReadDesignDocument("testing-a")
	assert.Nil(err)
	assert.Equal(ddocB.ID, ddocA.ID)
}

// TestUpdateDesignDocument tests updating design documents.
func TestUpdateDesignDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("update-design", assert)
	defer cleanup()

	// Create design document and read it again.
	ddocA := &couchdb.DesignDocument{
		ID: "_design/testing-a",
		Views: couchdb.DesignViews{
			"index-a": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('a') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp := cdb.CreateDesignDocument(ddocA)
	assert.True(resp.IsOK())

	ddocB, err := cdb.ReadDesignDocument("testing-a")
	assert.Nil(err)
	assert.Equal(ddocB.ID, ddocA.ID)

	// Now update it and read it again.
	ddocB.Views["index-b"] = couchdb.DesignView{
		Map: "function(doc){ if (doc._id.indexOf('b') !== -1) { emit(doc._id, doc._rev);  } }",
	}
	resp = cdb.UpdateDesignDocument(ddocB)
	assert.True(resp.IsOK())

	ddocC, err := cdb.ReadDesignDocument("testing-a")
	assert.Nil(err)
	assert.Length(ddocC.Views, 2)
}

// TestDeleteDesignDocument tests deleting design documents.
func TestDeleteDesignDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("delete-design", assert)
	defer cleanup()

	// Create design document and check if it has been created.
	allDesignA, err := cdb.AllDesignDocuments()
	assert.Nil(err)

	ddocA := &couchdb.DesignDocument{
		ID: "_design/testing-a",
		Views: couchdb.DesignViews{
			"index-a": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('a') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp := cdb.CreateDesignDocument(ddocA)
	assert.True(resp.IsOK())

	allDesignB, err := cdb.AllDesignDocuments()
	assert.Nil(err)
	assert.Equal(len(allDesignB), len(allDesignA)+1)

	// Read it and delete it.
	ddocB, err := cdb.ReadDesignDocument("testing-a")
	assert.Nil(err)

	resp = cdb.DeleteDesignDocument(ddocB)
	assert.True(resp.IsOK())

	allDesignC, err := cdb.AllDesignDocuments()
	assert.Nil(err)
	assert.Equal(len(allDesignC), len(allDesignA))
}

// TestViewDocuments tests calling a view.
func TestViewDocuments(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("view-documents", assert)
	defer cleanup()

	// Create design document.
	ddocA := &couchdb.DesignDocument{
		ID: "_design/testing",
		Views: couchdb.DesignViews{
			"index-a": couchdb.DesignView{
				Map: "function(doc){ if (doc._id.indexOf('a') !== -1) { emit(doc._id, doc._rev);  } }",
			},
		},
	}
	resp := cdb.CreateDesignDocument(ddocA)
	assert.True(resp.IsOK())

	// Call the view for the first time.
	resp = cdb.ViewDocuments("testing", "index-a")
	assert.True(resp.IsOK())
	vr := couchdb.ViewResult{}
	err := resp.ResultValue(&vr)
	assert.Nil(err)
	trOld := vr.TotalRows
	assert.True(trOld > 0)

	// Add a matching document and view again.
	docA := MyDocument{
		DocumentID: "black-jack-4711",
		Name:       "Jack Black",
	}
	resp = cdb.CreateDocument(docA)
	assert.True(resp.IsOK())
	resp = cdb.ViewDocuments("testing", "index-a")
	assert.True(resp.IsOK())
	vr = couchdb.ViewResult{}
	err = resp.ResultValue(&vr)
	assert.Nil(err)
	trNew := vr.TotalRows
	assert.Equal(trNew, trOld+1)

	// Add a non-matching document and view again.
	docB := MyDocument{
		DocumentID: "doe-john-999",
		Name:       "John Doe",
	}
	resp = cdb.CreateDocument(docB)
	assert.True(resp.IsOK())
	resp = cdb.ViewDocuments("testing", "index-a")
	assert.True(resp.IsOK())
	vr = couchdb.ViewResult{}
	err = resp.ResultValue(&vr)
	assert.Nil(err)
	trFinal := vr.TotalRows
	assert.Equal(trFinal, trNew)

}

// TestCreateDocument tests creating new documents.
func TestCreateDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareDatabase("create-document", assert)
	defer cleanup()

	// Create document without ID.
	docA := MyDocument{
		Name: "foo",
		Age:  50,
	}
	resp := cdb.CreateDocument(docA)
	assert.True(resp.IsOK())
	id := resp.ID()
	assert.Match(id, "[0-9a-f]{32}")

	// Create document with ID.
	docB := MyDocument{
		DocumentID: "bar-12345",
		Name:       "bar",
		Age:        25,
		Active:     true,
	}
	resp = cdb.CreateDocument(docB)
	assert.True(resp.IsOK())
	id = resp.ID()
	assert.Equal(id, "bar-12345")
}

// TestReadDocument tests reading a document.
func TestReadDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareDatabase("read-document", assert)
	defer cleanup()

	// Create test document.
	docA := MyDocument{
		DocumentID: "foo-12345",
		Name:       "foo",
		Age:        18,
	}
	resp := cdb.CreateDocument(docA)
	assert.True(resp.IsOK())
	id := resp.ID()
	assert.Equal(id, "foo-12345")

	// Read test document.
	resp = cdb.ReadDocument(id)
	assert.True(resp.IsOK())
	docB := MyDocument{}
	err := resp.ResultValue(&docB)
	assert.Nil(err)
	assert.Equal(docB.DocumentID, docA.DocumentID)
	assert.Equal(docB.Name, docA.Name)
	assert.Equal(docB.Age, docA.Age)

	// Try to read non-existant document.
	resp = cdb.ReadDocument("i-do-not-exist")
	assert.False(resp.IsOK())
	assert.ErrorMatch(resp.Error(), ".* 404,.*")
}

// TestUpdateDocument tests updating documents.
func TestUpdateDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareDatabase("update-document", assert)
	defer cleanup()

	// Create first revision.
	docA := MyDocument{
		DocumentID: "foo-12345",
		Name:       "foo",
		Age:        22,
	}
	resp := cdb.CreateDocument(docA)
	assert.True(resp.IsOK())
	id := resp.ID()
	revision := resp.Revision()
	assert.Equal(id, "foo-12345")

	resp = cdb.ReadDocument(id)
	assert.True(resp.IsOK())
	docB := MyDocument{}
	err := resp.ResultValue(&docB)
	assert.Nil(err)

	// Update the document.
	docB.Age = 23

	resp = cdb.UpdateDocument(docB)
	assert.True(resp.IsOK())

	// read the updated revision.
	resp = cdb.ReadDocument(id)
	assert.True(resp.IsOK())
	docC := MyDocument{}
	err = resp.ResultValue(&docC)
	assert.Nil(err)
	assert.Equal(docC.DocumentID, docB.DocumentID)
	assert.Substring("2-", docC.DocumentRevision)
	assert.Equal(docC.Name, docB.Name)
	assert.Equal(docC.Age, docB.Age)

	// Read the first revision.
	resp = cdb.ReadDocument(id, couchdb.Revision(revision))
	assert.True(resp.IsOK())
	docD := MyDocument{}
	err = resp.ResultValue(&docD)
	assert.Nil(err)
	assert.Equal(docD.DocumentRevision, revision)
	assert.Equal(docD.Age, docA.Age)
}

// TestDeleteDocument tests deleting a document.
func TestDeleteDocument(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareDatabase("delete-document", assert)
	defer cleanup()

	// Create test document.
	docA := MyDocument{
		DocumentID: "foo-12345",
		Name:       "foo",
		Age:        33,
	}
	resp := cdb.CreateDocument(docA)
	assert.True(resp.IsOK())
	id := resp.ID()
	assert.Equal(id, "foo-12345")

	// Read test document, we need it including the revision.
	resp = cdb.ReadDocument(id)
	assert.True(resp.IsOK())
	docB := MyDocument{}
	err := resp.ResultValue(&docB)
	assert.Nil(err)

	// Delete the test document.
	resp = cdb.DeleteDocument(docB)
	assert.True(resp.IsOK())

	// Try to read deleted document.
	resp = cdb.ReadDocument(id)
	assert.False(resp.IsOK())
	assert.ErrorMatch(resp.Error(), ".* 404,.*")
}

//--------------------
// HELPERS
//--------------------

// MyDocument is used for the tests.
type MyDocument struct {
	DocumentID       string `json:"_id,omitempty"`
	DocumentRevision string `json:"_rev,omitempty"`

	Name        string `json:"name"`
	Age         int    `json:"age"`
	Active      bool   `json:"active"`
	Description string `json:"description"`
}

// prepareDatabase opens the database, deletes a possible test
// database, and creates it newly.
func prepareDatabase(database string, assert audit.Assertion) (couchdb.CouchDB, func()) {
	cfgstr := strings.Replace(TemplateDBcfg, "<<DATABASE>>", database, 1)
	cfg, err := etc.ReadString(cfgstr)
	assert.Nil(err)
	cdb, err := couchdb.Open(cfg)
	assert.Nil(err)
	resp := cdb.DeleteDatabase()
	resp = cdb.CreateDatabase()
	assert.True(resp.IsOK())
	return cdb, func() { cdb.DeleteDatabase() }
}

// prepareFilledDatabase opens the database, deletes a possible test
// database, creates it newly and adds some data.
func prepareFilledDatabase(database string, assert audit.Assertion) (couchdb.CouchDB, func()) {
	cfgstr := strings.Replace(TemplateDBcfg, "<<DATABASE>>", database, 1)
	cfg, err := etc.ReadString(cfgstr)
	assert.Nil(err)
	cdb, err := couchdb.Open(cfg)
	assert.Nil(err)
	resp := cdb.DeleteDatabase()
	resp = cdb.CreateDatabase()
	assert.True(resp.IsOK())

	gen := audit.NewGenerator(audit.FixedRand())
	docs := []interface{}{}
	for i := 0; i < 1000; i++ {
		first, middle, last := gen.Name()
		doc := MyDocument{
			DocumentID:  identifier.Identifier(last, first, i),
			Name:        first + " " + middle + " " + last,
			Age:         gen.Int(18, 65),
			Active:      gen.FlipCoin(75),
			Description: gen.Sentence(),
		}
		docs = append(docs, doc)
	}
	results, err := cdb.BulkWriteDocuments(docs...)
	assert.Nil(err)
	for _, result := range results {
		assert.True(result.OK)
	}

	return cdb, func() { cdb.DeleteDatabase() }
}

// EOF
