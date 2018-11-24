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
	"reflect"
	"testing"
	"time"
)

// Test getting properties and items from a struct

type SimpleTestStruct struct {
	Name    string    `property:"propname"`
	Address time.Time `property:"address"`
	Unused  string
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
	wikibase := NewClient(client)

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
	wikibase := NewClient(client)

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
	wikibase := NewClient(client)

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

	v, err := stringClaimToAPIData(testdata)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
	if v == nil {
	    t.Errorf("Expected non nil return")
	}
}

func TestZeroLengthStringClaimEncode(t *testing.T) {

	const testdata = ""

	v, err := stringClaimToAPIData(testdata)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
	if v != nil {
	    t.Errorf("Zero length string should return nil: %v", v)
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
	_, err := timeDataClaimToAPIData("1976-06-06T13:45:02Z")
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}
}

// Test marshalling of claims
type marshalTestStruct struct {
	A string
	B int
	C time.Time
	D ItemPropertyType
	E *int
	F *int
	G *time.Time
	H *ItemPropertyType
	I string
}

func TestMarshalInternal(t *testing.T) {

	a := 24
	b := ItemPropertyType("Q23")
	s := marshalTestStruct{
		A: "hello",
		B: 42,
		C: time.Now(),
		D: "Q43",
		E: nil,
		F: &a,
		G: nil,
		H: &b,
		I: "", // wikidata doesn't cope with zero length strings, so we should return no value for this
	}
	expectData := []bool{true, true, true, true, false, true, false, true, false}

	r := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		value := v.Field(i)

		data, err := getDataForClaim(field, value)
		if err != nil {
			t.Fatalf("Failed to marshal claim %d: %v", i, err)
		}
		if expectData[i] && data == nil {
			t.Fatalf("We got no data for field %d", i)
		}
	}
}
