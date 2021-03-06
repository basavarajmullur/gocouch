// Tideland Go CouchDB Client - Find - Unit Tests
//
// Copyright (C) 2016-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package find_test

import (
	"strings"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/etc"
	"github.com/tideland/golib/identifier"

	"github.com/tideland/gocouch/couchdb"
	"github.com/tideland/gocouch/find"
)

//--------------------
// IMPORTS
//--------------------

//--------------------
// CONSTANTS
//--------------------

const (
	Cfg = "{etc {hostname localhost}{port 5984}{database tgocouch-testing-<<DATABASE>>}{debug-logging true}}"
)

//--------------------
// TESTS
//--------------------

// TestSimpleFind tests calling find with a simple selector.
func TestSimpleFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-simple", 1000, assert)
	defer cleanup()

	// Try to find some documents a simple way.
	selector := find.Select(find.Or(
		find.And(
			find.LowerThan("age", 30),
			find.Equal("active", false),
		),
		find.And(
			find.GreaterThan("age", 60),
			find.Equal("active", "true"),
		),
	))
	frs := find.Find(cdb, selector, find.Fields("name", "age", "active"))
	assert.Nil(frs.Error())
	assert.True(frs.IsOK())

	err := frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name   string `json:"name"`
			Age    int    `json:"age"`
			Active bool   `json:"active"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.True((fields.Age < 30 && !fields.Active) || (fields.Age > 60 && fields.Active))
		return nil
	})
	assert.Nil(err)
}

// TestLimitedFind tests retrieving a larger number but set the limit.
func TestLimitedFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-limited", 1000, assert)
	defer cleanup()

	// Limit found documents.
	selector := find.Select(find.Equal("active", true))
	frs := find.Find(cdb, selector, find.Fields("name", "active"), find.Limit(5))
	assert.NotNil(frs)
	assert.True(frs.IsOK())
	assert.Length(frs, 5)

	// Greater limit.
	frs = find.Find(cdb, selector, find.Fields("name", "active"), find.Limit(100))
	assert.NotNil(frs)
	assert.True(frs.IsOK())
	assert.Length(frs, 100)
}

// TestSortedFind tests retrieving a larger number in a sorted way.
func TestSortedFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-sorted", 1000, assert)
	defer cleanup()

	// Sorting field has to be part of selector.
	selector := find.Select(
		find.GreaterThan("name", nil),
		find.Equal("active", true),
	)
	frs := find.Find(cdb, selector, find.Fields("name", "age"), find.Sort(find.Ascending("name")), find.Limit(1000))
	assert.NotNil(frs)
	assert.Equal(frs.Error(), nil)
	assert.True(frs.IsOK())

	name := ""
	err := frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.True(fields.Name >= name)
		name = fields.Name
		return nil
	})
	assert.Nil(err)
}

// TestFindExists tests calling find with an exists selector.
func TestFindExists(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-exists", 1000, assert)
	defer cleanup()

	// Try to find some documents having an existing "last_active".
	selector := find.Select(
		find.Exists("last_active"),
		find.LowerEqualThan("age", 25),
	)
	frs := find.Find(cdb, selector, find.Fields("name", "age", "active", "last_active"))
	assert.NotNil(frs)
	assert.True(frs.IsOK())

	err := frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name       string `json:"name"`
			Age        int    `json:"age"`
			Active     bool   `json:"active"`
			LastActive int64  `json:"last_active"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.True(fields.Age <= 25 && fields.LastActive > 0 && fields.Active)
		return nil
	})
	assert.Nil(err)

	// Now look for existing "last_active" but "active" is false. So
	// no results.
	selector = find.Select(
		find.Exists("last_active"),
		find.NotEqual("active", true),
	)
	frs = find.Find(cdb, selector, find.Fields("name", "age", "active", "last_active"))
	assert.NotNil(frs)
	assert.True(frs.IsOK())
	assert.Equal(frs.Len(), 0)
}

// TestOneCriterion tests using only one criterion, here a regular expression.
func TestOneCriterion(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-one-criterion", 1000, assert)
	defer cleanup()

	// Try to find some documents having an existing "last_active".
	selector := find.Select(find.RegEx("name", ".*Adam.*"))
	frs := find.Find(cdb, selector, find.Fields("name", "age", "active"))
	assert.NotNil(frs)
	assert.True(frs.IsOK())

	err := frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name   string `json:"name"`
			Age    int    `json:"age"`
			Active bool   `json:"active"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.Match(fields.Name, ".*Adam.*")
		return nil
	})
	assert.Nil(err)
}

// TestMatches tests using element and all match operators.
func TestMatches(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	cdb, cleanup := prepareFilledDatabase("find-match", 1000, assert)
	defer cleanup()

	// Find with at least one matching element.
	selector := find.Select(find.MatchElement("shifts", find.Equal("", 3)))
	frs := find.Find(cdb, selector, find.Fields("name", "shifts"))
	assert.NotNil(frs)
	assert.True(frs.IsOK())

	err := frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name   string `json:"name"`
			Shifts []int  `json:"shifts"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.Contents(3, fields.Shifts)
		return nil
	})
	assert.Nil(err)

	// Find with all matching elements (dumb query,
	// but checking combination).
	selector = find.Select(
		find.MatchAll("shifts",
			find.GreaterThan("", 1),
			find.LowerThan("", 3),
		),
	)
	frs = find.Find(cdb, selector, find.Fields("name", "shifts"))
	assert.NotNil(frs)
	assert.Nil(frs.Error())
	assert.True(frs.IsOK())

	err = frs.Do(func(document couchdb.Unmarshable) error {
		fields := struct {
			Name   string `json:"name"`
			Shifts []int  `json:"shifts"`
		}{}
		if err := document.Unmarshal(&fields); err != nil {
			return err
		}
		assert.Equal(fields.Shifts, []int{2, 2, 2})
		return nil
	})
	assert.Nil(err)
}

//--------------------
// HELPERS
//--------------------

// Note is used for the tests.
type Note struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Person is used for the tests.
type Person struct {
	DocumentID       string `json:"_id,omitempty"`
	DocumentRevision string `json:"_rev,omitempty"`

	Name       string `json:"name"`
	Age        int    `json:"age"`
	Shifts     []int  `json:"shifts"`
	Active     bool   `json:"active"`
	LastActive int64  `json:"last_active,omitempty"`
	Notes      []Note `json:"notes"`
}

// prepareFilledDatabase opens the database, deletes a possible test
// database, creates it newly and adds some data.
func prepareFilledDatabase(database string, count int, assert audit.Assertion) (couchdb.CouchDB, func()) {
	cfgstr := strings.Replace(Cfg, "<<DATABASE>>", database, 1)
	cfg, err := etc.ReadString(cfgstr)
	assert.Nil(err)
	cdb, err := couchdb.Open(cfg)
	assert.Nil(err)
	rs := cdb.DeleteDatabase()
	rs = cdb.CreateDatabase()
	assert.True(rs.IsOK())
	err = find.CreateIndex(cdb, find.NewIndex("name"))
	assert.Nil(err)

	gen := audit.NewGenerator(audit.FixedRand())
	runs := count / 1000
	for outer := 0; outer < runs; outer++ {
		assert.Logf("filling database run %d of %d", outer+1, runs)
		docs := []interface{}{}
		for inner := 0; inner < 1000; inner++ {
			first, middle, last := gen.Name()
			person := Person{
				DocumentID: identifier.Identifier(last, first, outer, inner),
				Name:       first + " " + middle + " " + last,
				Age:        gen.Int(18, 65),
				Shifts:     []int{gen.Int(1, 3), gen.Int(1, 3), gen.Int(1, 3)},
				Active:     gen.FlipCoin(75),
			}
			if person.Active {
				person.LastActive = gen.Time(time.UTC, time.Now().Add(-24*time.Hour), 24*time.Hour).Unix()
			}
			for j := 0; j < gen.Int(3, 9); j++ {
				note := Note{
					Title: gen.Sentence(),
					Text:  gen.Paragraph(),
				}
				person.Notes = append(person.Notes, note)
			}
			docs = append(docs, person)
		}
		results, err := cdb.BulkWriteDocuments(docs)
		assert.Nil(err)
		for _, result := range results {
			assert.True(result.OK)
		}
	}

	return cdb, func() { cdb.DeleteDatabase() }
}

// EOF
