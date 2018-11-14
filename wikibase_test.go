//   Copyright 2018 Content Mine Ltd
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package wikibase

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

// Test network layer substitute

type WikiBaseNetworkTestClient struct {
	Error    error
	Response string
	Args     map[string]string
}

func (c *WikiBaseNetworkTestClient) Get(args map[string]string) (io.ReadCloser, error) {

	c.Args = args

	if c.Error != nil {
		return nil, c.Error
	}

	return ioutil.NopCloser(bytes.NewBuffer([]byte(c.Response))), nil
}

func (c *WikiBaseNetworkTestClient) Post(args map[string]string) (io.ReadCloser, error) {

	c.Args = args

	if c.Error != nil {
		return nil, c.Error
	}

	return ioutil.NopCloser(bytes.NewBuffer([]byte(c.Response))), nil
}

func createClientWithResponse(str string) *WikiBaseNetworkTestClient {
	return &WikiBaseNetworkTestClient{Response: str}
}

func createClientWithError(err error) *WikiBaseNetworkTestClient {
	return &WikiBaseNetworkTestClient{Error: err}
}

// Actual tests

func TestErrorGettingEditingToken(t *testing.T) {

	client := createClientWithError(fmt.Errorf("Oops"))
	wikibase := NewWikiBaseClient(client)

	_, err := wikibase.GetEditingToken()

	if err == nil {
		t.Errorf("Expected an error but didn't get one")
	}

	// Check that the request was also sane
	if client.Args["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
	if client.Args["meta"] != "tokens" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
}

func TestErrorGettingEditingTokenWhenAlreadyExists(t *testing.T) {

	client := createClientWithError(fmt.Errorf("Oops"))
	wikibase := NewWikiBaseClient(client)
	token := "inserttokenhere"
	wikibase.editToken = &token

	resp, err := wikibase.GetEditingToken()

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != token {
	    t.Errorf("Got unexpected token: %s", resp)
	}

	// Check that the request wasn't made
	if len(client.Args) != 0 {
		t.Errorf("Unexpected args requested: %v", client.Args)
	}
}

func TestGettingEditingToken(t *testing.T) {

	client := createClientWithResponse(`
{"batchcomplete":"","query":{"tokens":{"csrftoken":"345def4e73a103a0ea37f924f999ffad5be05458+\\\\"}}}
`)
	wikibase := NewWikiBaseClient(client)

	resp, err := wikibase.GetEditingToken()

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != "345def4e73a103a0ea37f924f999ffad5be05458+\\\\" {
		t.Errorf("Token did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.Args["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
	if client.Args["meta"] != "tokens" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
}

func TestGettingItemForLabel(t *testing.T) {

	client := createClientWithResponse(`
{
    "batchcomplete": "",
    "requestid": "42",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Item:Q4",
                "pageid": 11,
                "displaytext": "wibble"
            }
        ]
    }
}
`)
	wikibase := NewWikiBaseClient(client)

	resp, err := wikibase.GetItemForLabel("blah")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != "Q4" {
		t.Errorf("ID did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.Args["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
	if client.Args["list"] != "wbsearch" {
		t.Errorf("Unexpected list requested: %v", client.Args)
	}
	if client.Args["wbssearch"] != "blah" {
		t.Errorf("Unexpected search requested: %v", client.Args)
	}
	if client.Args["wbstype"] != "item" {
		t.Errorf("Unexpected type requested: %v", client.Args)
	}
}

func TestGettingPropertyForLabel(t *testing.T) {

	client := createClientWithResponse(`
{
    "batchcomplete": "",
    "requestid": "42",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Property:P25",
                "pageid": 11,
                "displaytext": "wibble"
            }
        ]
    }
}
`)
	wikibase := NewWikiBaseClient(client)

	resp, err := wikibase.GetPropertyForLabel("blah")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != "P25" {
		t.Errorf("ID did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.Args["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
	if client.Args["list"] != "wbsearch" {
		t.Errorf("Unexpected list requested: %v", client.Args)
	}
	if client.Args["wbssearch"] != "blah" {
		t.Errorf("Unexpected search requested: %v", client.Args)
	}
	if client.Args["wbstype"] != "property" {
		t.Errorf("Unexpected type requested: %v", client.Args)
	}
}

func TestCreateItem(t *testing.T) {

	client := createClientWithResponse(`
{
    "entity": {
        "aliases": {},
        "claims": {},
        "descriptions": {},
        "id": "Q11",
        "labels": {
            "en": {
                "language": "en",
                "value": "hello"
            }
        },
        "lastrevid": 55,
        "sitelinks": {},
        "type": "item"
    },
    "success": 1
}
`)
	wikibase := NewWikiBaseClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token

	resp, err := wikibase.CreateItemInstance("blah")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != "Q11" {
		t.Errorf("ID did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.Args["action"] != "wbeditentity" {
		t.Errorf("Unexpected action requested: %v", client.Args)
	}
	if client.Args["token"] != token {
		t.Errorf("Unexpected token requested: %v", client.Args)
	}
	if client.Args["new"] != "item" {
		t.Errorf("Unexpected search requested: %v", client.Args)
	}
}

func TestCreateItemWithoutEditToken(t *testing.T) {

	client := createClientWithResponse(`
{
    "entity": {
        "aliases": {},
        "claims": {},
        "descriptions": {},
        "id": "Q11",
        "labels": {
            "en": {
                "language": "en",
                "value": "hello"
            }
        },
        "lastrevid": 55,
        "sitelinks": {},
        "type": "item"
    },
    "success": 1
}
`)
	wikibase := NewWikiBaseClient(client)

	_, err := wikibase.CreateItemInstance("blah")

	if err == nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}
