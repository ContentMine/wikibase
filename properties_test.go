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
    "fmt"
	"testing"
	"time"
	"unicode/utf8"
)


// Test getting properties and items from a struct

type SimpleTestStruct struct {
    Name string `property:"propname"`
    Address string `property:"address"`
    Unused string
}

func TestParseSimpleStruct(t *testing.T) {

    client := &WikiBaseNetworkTestClient{}
    client.addDataResponse(`
{
    "batchcomplete": "",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Property:P23",
                "pageid": 11,
                "displaytext": "propname"
            }
        ]
    }
}
`)
    client.addDataResponse(`
{
    "batchcomplete": "",
    "query": {
        "wbsearch": [
            {
                "ns": 120,
                "title": "Property:P5",
                "pageid": 11,
                "displaytext": "address"
            }
        ]
    }
}
`)
	wikibase := NewWikiBaseClient(client)

    err := wikibase.MapPropertyAndItemConfiguration(SimpleTestStruct{})
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }

    if len(wikibase.PropertyMap) != 2 {
        t.Fatalf("Our property map does not have enough items: %v", wikibase.PropertyMap)
    }
}

func TestParseSimpleStructErrors(t *testing.T) {

    client := &WikiBaseNetworkTestClient{}
    client.addErrorResponse(fmt.Errorf("Oops"))
	wikibase := NewWikiBaseClient(client)

    err := wikibase.MapPropertyAndItemConfiguration(SimpleTestStruct{})
    if err == nil {
        t.Fatalf("We expected an error")
    }
}

func TestMapItemByName(t *testing.T) {

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
	wikibase := NewWikiBaseClient(client)

	err := wikibase.MapItemConfigurationByLabel("blah")
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }

    if len(wikibase.ItemMap) != 1 {
        t.Fatalf("Our item map does not have enough items: %v", wikibase.ItemMap)
    }
}

// Tests for API Encoding of claims

func TestStringClaimEncode(t *testing.T) {

	const testdata = "hello, world"

	data, err := stringClaimToAPIData(testdata)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
	// A cheap test, but check that the returned string is two characters longer due to being in quotes
	if utf8.RuneCountInString(testdata)+2 != utf8.RuneCountInString(string(data)) {
		t.Fatalf("Length of encoded data not what we expected: %v", string(data))
	}
}

func TestItemClaimEncode(t *testing.T) {
	_, err := itemClaimToAPIData("Q42")
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
}

func TestPropertyAsItemClaimEncode(t *testing.T) {
	_, err := itemClaimToAPIData("P42")
	if err == nil {
		t.Fatalf("We got an expected an error")
	}
}

func TestInvalidItemClaimEncode(t *testing.T) {
	_, err := itemClaimToAPIData("42")
	if err == nil {
		t.Fatalf("We got an expected an error")
	}
}

func TestQuntityClaimEncode(t *testing.T) {
	_, err := quantityClaimToAPIData(42)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
}

func TestTimeDataClaimEncode(t *testing.T) {
	_, err := timeDataClaimToAPIData(time.Now())
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
}
