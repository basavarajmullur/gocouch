// Tideland Go CouchDB Client - Security - Unit Tests
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package security_test

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/etc"

	"github.com/tideland/gocouch/couchdb"
	"github.com/tideland/gocouch/security"
)

//--------------------
// CONSTANTS
//--------------------

const (
	TemplateDBcfg = "{etc {hostname localhost}{port 5984}{database tgocouch-testing-<<DATABASE>>}}"
)

//--------------------
// TESTS
//--------------------

// TestWriteDeleteAdministrator tests the creation of the initial
// and a second administrator and also the deletion of them.
func TestWriteDeleteAdministrator(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb := prepareCouchDB("write-delete-administrator", assert)

	err := security.WriteAdministrator(cdb, nil, "admin1", "admin1")
	assert.Nil(err)
	defer func() {
		// Let the administator remove himself.
		session, err := security.NewSession(cdb, "admin1", "admin1")
		assert.Nil(err)
		err = security.DeleteAdministrator(cdb, session, "admin1")
		assert.Nil(err)
	}()

	session, err := security.NewSession(cdb, "admin1", "admin1")
	assert.Nil(err)
	err = security.WriteAdministrator(cdb, session, "admin2", "admin2")
	assert.Nil(err)
	err = security.DeleteAdministrator(cdb, session, "admin2")
	assert.Nil(err)
}

// TestHasAdministrator tests the check for an administrator.
func TestHasAdministrator(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb := prepareCouchDB("has-administrator", assert)

	ok, err := security.HasAdministrator(cdb, nil, "admin")
	assert.Nil(err)
	assert.False(ok)

	err = security.WriteAdministrator(cdb, nil, "admin", "admin")
	assert.Nil(err)
	defer func() {
		// Let the administator remove himself.
		session, err := security.NewSession(cdb, "admin", "admin")
		assert.Nil(err)
		err = security.DeleteAdministrator(cdb, session, "admin")
		assert.Nil(err)
	}()

	ok, err = security.HasAdministrator(cdb, nil, "admin")
	assert.ErrorMatch(err, ".*status code 401.*")
	assert.False(ok)

	session, err := security.NewSession(cdb, "admin", "admin")
	assert.Nil(err)
	ok, err = security.HasAdministrator(cdb, session, "admin")
	assert.Nil(err)
	assert.True(ok)
}

// TestWriteAdministratorNoSession tests the creation of another
// admin if the creator has no valid session.
func TestWriteAdministratorNoSession(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb := prepareCouchDB("write-administrator-no-session", assert)

	err := security.WriteAdministrator(cdb, nil, "admin1", "admin1")
	assert.Nil(err)
	defer func() {
		// Let the administator remove himself.
		session, err := security.NewSession(cdb, "admin1", "admin1")
		assert.Nil(err)
		err = security.DeleteAdministrator(cdb, session, "admin1")
		assert.Nil(err)
	}()

	err = security.WriteAdministrator(cdb, nil, "admin2", "admin2")
	assert.ErrorMatch(err, ".*status code 401.*")
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

// prepareCouchDB opens the DBMS for one database
// w/o creating it. It deletes the named database to
// avoid conflicts.
func prepareCouchDB(database string, assert audit.Assertion) couchdb.CouchDB {
	cfgstr := strings.Replace(TemplateDBcfg, "<<DATABASE>>", database, 1)
	cfg, err := etc.ReadString(cfgstr)
	assert.Nil(err)
	cdb, err := couchdb.Open(cfg)
	assert.Nil(err)
	cdb.DeleteDatabase()
	return cdb
}

// EOF
