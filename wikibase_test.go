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

type WikiBaseNetworkTestClientResponse struct {
	Data  string
	Error error
}

type WikiBaseNetworkTestClient struct {
	InvocationCount int
	Responses       []WikiBaseNetworkTestClientResponse
	MostRecentArgs  map[string]string
}

func (c *WikiBaseNetworkTestClient) innerCall(args map[string]string) (io.ReadCloser, error) {
	c.MostRecentArgs = args

	resp := c.Responses[c.InvocationCount]
	c.InvocationCount += 1

	if resp.Error != nil {
		return nil, resp.Error
	}

	return ioutil.NopCloser(bytes.NewBuffer([]byte(resp.Data))), nil
}

func (c *WikiBaseNetworkTestClient) Get(args map[string]string) (io.ReadCloser, error) {
	return c.innerCall(args)
}

func (c *WikiBaseNetworkTestClient) Post(args map[string]string) (io.ReadCloser, error) {
	return c.innerCall(args)
}

func (c *WikiBaseNetworkTestClient) addDataResponse(data string) {
	if c.Responses == nil {
		c.Responses = make([]WikiBaseNetworkTestClientResponse, 0)
	}
	c.Responses = append(c.Responses, WikiBaseNetworkTestClientResponse{Data: data})
}

func (c *WikiBaseNetworkTestClient) addErrorResponse(err error) {
	if c.Responses == nil {
		c.Responses = make([]WikiBaseNetworkTestClientResponse, 0)
	}
	c.Responses = append(c.Responses, WikiBaseNetworkTestClientResponse{Error: err})
}

// Actual tests

func TestErrorGettingEditingToken(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addErrorResponse(fmt.Errorf("Oops"))

	wikibase := NewClient(client)

	_, err := wikibase.GetEditingToken()

	if err == nil {
		t.Errorf("Expected an error but didn't get one")
	}

	// Check that the request was also sane
	if client.MostRecentArgs["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["meta"] != "tokens" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
}

func TestErrorGettingEditingTokenWhenAlreadyExists(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addErrorResponse(fmt.Errorf("Oops"))

	wikibase := NewClient(client)
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
	if len(client.MostRecentArgs) != 0 {
		t.Errorf("Unexpected args requested: %v", client.MostRecentArgs)
	}
}

func TestGettingEditingToken(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"batchcomplete":"","query":{"tokens":{"csrftoken":"345def4e73a103a0ea37f924f999ffad5be05458+\\\\"}}}
`)
	wikibase := NewClient(client)

	resp, err := wikibase.GetEditingToken()

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if resp != "345def4e73a103a0ea37f924f999ffad5be05458+\\\\" {
		t.Errorf("Token did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.MostRecentArgs["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["meta"] != "tokens" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
}

func TestGettingItemForLabel(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{
    "batchcomplete": "",
    "requestid": "42",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Item:Q4",
                "pageid": 11,
                "displaytext": "blah"
            }
        ]
    }
}
`)
	wikibase := NewClient(client)

	resp, err := wikibase.FetchItemIDsForLabel("blah")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Got more response than expected: %v", resp)
	}
	if resp[0] != "Q4" {
		t.Errorf("ID did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.MostRecentArgs["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["list"] != "wbsearch" {
		t.Errorf("Unexpected list requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["wbssearch"] != "blah" {
		t.Errorf("Unexpected search requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["wbstype"] != "item" {
		t.Errorf("Unexpected type requested: %v", client.MostRecentArgs)
	}
}

func TestGettingUniqueItemForLabel(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
    	{"batchcomplete":"","query":{"wbsearch":[{"ns":120,"title":"Item:Q6","pageid":33,"displaytext":"annotation"},{"ns":120,"title":"Item:Q101","pageid":128,"displaytext":"annotation instance"},{"ns":120,"title":"Item:Q103","pageid":130,"displaytext":"annotation instance"},{"ns":120,"title":"Item:Q105","pageid":132,"displaytext":"annotation instance"},{"ns":120,"title":"Item:Q107","pageid":134,"displaytext":"annotation instance"},{"ns":120,"title":"Item:Q109","pageid":136,"displaytext":"annotation instance"},{"ns":120,"title":"Item:Q111","pageid":138,"displaytext":"annotation instance"}]}}
`)
	wikibase := NewClient(client)

	resp, err := wikibase.FetchItemIDsForLabel("annotation")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Got more response than expected: %v", resp)
	}
	if resp[0] != "Q6" {
		t.Errorf("ID did not match expected: %s", resp)
	}
}

func TestGettingPropertyForLabel(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{
    "batchcomplete": "",
    "requestid": "42",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Property:P25",
                "pageid": 11,
                "displaytext": "blah"
            }
        ]
    }
}
`)
	wikibase := NewClient(client)

	resp, err := wikibase.FetchPropertyIDsForLabel("blah")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Got more response than expected: %v", resp)
	}
	if resp[0] != "P25" {
		t.Errorf("ID did not match expected: %s", resp)
	}

	// Check that the request was also sane
	if client.MostRecentArgs["action"] != "query" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["list"] != "wbsearch" {
		t.Errorf("Unexpected list requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["wbssearch"] != "blah" {
		t.Errorf("Unexpected search requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["wbstype"] != "property" {
		t.Errorf("Unexpected type requested: %v", client.MostRecentArgs)
	}
}

func TestCreateItem(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
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
	wikibase := NewClient(client)
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
	if client.MostRecentArgs["action"] != "wbeditentity" {
		t.Errorf("Unexpected action requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["token"] != token {
		t.Errorf("Unexpected token requested: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["new"] != "item" {
		t.Errorf("Unexpected search requested: %v", client.MostRecentArgs)
	}
}

func TestCreateItemWithoutEditToken(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
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
	wikibase := NewClient(client)

	_, err := wikibase.CreateItemInstance("blah")

	if err == nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}

// Page protection tests

func TestProtectPageByID(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
    	{"protect":{"title":"Hello","reason":"","protections":[{"edit":"sysop","expiry":"infinite"}]}}
`)
	wikibase := NewClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token

	err := wikibase.ProtectPageByID(42)

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}

func TestProtectPageByTitle(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
    	{"protect":{"title":"Hello","reason":"","protections":[{"edit":"sysop","expiry":"infinite"}]}}
`)
	wikibase := NewClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token

	err := wikibase.ProtectPageByTitle("hello")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}

func TestProtectPageGetsError(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
    	 {"error":{"code":"nosuchpageid","info":"There is no page with ID 742232.","*":"See http://localhost:8181/w/api.php for API usage. Subscribe to the mediawiki-api-announce mailing list at &lt;https://lists.wikimedia.org/mailman/listinfo/mediawiki-api-announce&gt; for notice of API deprecations and breaking changes."}}
`)
	wikibase := NewClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token

	err := wikibase.ProtectPageByID(42)

	if err == nil {
		t.Errorf("We expected an error")
	}
}
