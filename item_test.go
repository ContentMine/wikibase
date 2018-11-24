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
	"strings"
	"testing"
)

type SimpleItemTestStruct struct {
	ItemHeader
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

	item := SimpleItemTestStruct{}
	err := wikibase.CreateItemInstance("blah", &item)

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if item.ID != "Q11" {
		t.Errorf("ID did not match expected: %s", item)
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

	item := SimpleItemTestStruct{}
	err := wikibase.CreateItemInstance("blah", &item)

	if err == nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}

type SingleClaimTestStruct struct {
	ItemHeader

	Test string `property:"test"`
}

func TestCreateItemWithProperty(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{
    "entity": {
        "aliases": {},
        "claims": {
            "P19": [
                {
                    "id": "Q7924$A3F81E52-23FF-4284-8076-E6BF2523C409",
                    "mainsnak": {
                        "datatype": "string",
                        "datavalue": {
                            "type": "string",
                            "value": "wibble"
                        },
                        "hash": "9232e7703e5b44d84d4ff9a1f03c2839d8c47f17",
                        "property": "P19",
                        "snaktype": "value"
                    },
                    "rank": "normal",
                    "type": "statement"
                }
            ]
        },
        "descriptions": {},
        "id": "Q7924",
        "labels": {
            "en": {
                "language": "en",
                "value": "foo"
            }
        },
        "lastrevid": 78256,
        "sitelinks": {},
        "type": "item"
    },
    "success": 1
}
`)
	wikibase := NewClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token
	wikibase.PropertyMap["test"] = "P19"

	item := SingleClaimTestStruct{Test: "wibble"}
	err := wikibase.CreateItemInstance("blah", &item)

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if item.ID != "Q7924" {
		t.Errorf("ID did not match expected: %s", item)
	}
	if len(item.PropertyIDs) != 1 {
		t.Fatalf("Property map does not contain expected values: %v", item)
	}
	if item.PropertyIDs["P19"] != "Q7924$A3F81E52-23FF-4284-8076-E6BF2523C409" {
		t.Errorf("Property map has wrong properties set: %v", item.PropertyIDs["P19"])
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
	if strings.Index(client.MostRecentArgs["data"], "wibble") == -1 {
		t.Errorf("Failed to spot test data in API call: %v", client.MostRecentArgs)
	}
}

func TestUploadClaim(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if len(item.PropertyIDs) != 1 {
		t.Fatalf("We expected to have stored a property ID: %v", item)
	}
	if item.PropertyIDs["P14"] != "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63" {
		t.Errorf("We got the wrong property ID: %v", item.PropertyIDs)
	}
}

func TestUploadClaimWithInitialisedMap(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"
	item.PropertyIDs = make(map[string]string, 0)

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if len(item.PropertyIDs) != 1 {
		t.Fatalf("We expected to have stored a property ID: %v", item)
	}
	if item.PropertyIDs["P14"] != "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63" {
		t.Errorf("We got the wrong property ID: %v", item.PropertyIDs)
	}
}

func TestUploadClaimWithExistingProperty(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"
	item.PropertyIDs = make(map[string]string, 0)
	item.PropertyIDs["P14"] = "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63"

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if client.InvocationCount != 0 {
		t.Errorf("Got unexpected invocation count: %v", client)
	}
}

func TestUploadClaimWithExistingPropertyButAllowRefresh(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"
	item.PropertyIDs = make(map[string]string, 0)
	item.PropertyIDs["P14"] = "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63"

	err := wikibase.UploadClaimsForItem(&item, true)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if client.InvocationCount != 1 {
		t.Errorf("Got unexpected invocation count: %v", client)
	}
}

func TestUploadClaimWithoutPointer(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(item, false)
	if err == nil {
		t.Fatalf("We expected an error")
	}
}

func TestUploadClaimWithArrayItem(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	items := make([]SingleClaimTestStruct, 1)

	items[0].Test = "blah"
	items[0].ID = "Q23"

	err := wikibase.UploadClaimsForItem(items[0], false)
	if err == nil {
		t.Fatalf("We expected an error")
	}
}

func TestUploadClaimWithArrayItemPointer(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	items := make([]SingleClaimTestStruct, 1)

	items[0].Test = "blah"
	items[0].ID = "Q23"

	err := wikibase.UploadClaimsForItem(&items[0], false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if len(items[0].PropertyIDs) != 1 {
		t.Fatalf("We expected to have stored a property ID: %v", items[0])
	}
	if items[0].PropertyIDs["P14"] != "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63" {
		t.Errorf("We got the wrong property ID: %v", items[0].PropertyIDs)
	}
}

type PointerPropertyClaimTestStruct struct {
	ItemHeader

	Test *string `property:"test"`
}

func TestUploadClaimNilPointer(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	// nil structure
	item := PointerPropertyClaimTestStruct{}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if client.MostRecentArgs["snaktype"] != "novalue" {
		t.Errorf("We got unexpected arguments for nil property: %v", client.MostRecentArgs)
	}
}

func TestUploadClaimValidPointer(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	a := "foo"
	item := PointerPropertyClaimTestStruct{Test: &a}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if client.MostRecentArgs["snaktype"] != "value" {
		t.Errorf("We got unexpected snaktype argument for non-nil property: %v", client.MostRecentArgs)
	}
	if client.MostRecentArgs["value"] != "\"foo\"" {
		t.Errorf("We got unexpected value argument for non-nil property: %v", client.MostRecentArgs)
	}
}

type SingleClaimWithoutInitialUploadTestStruct struct {
	ItemHeader

	Test string `property:"test,omitoncreate"`
}

func TestCreateItemWithOmitProperty(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{
    "entity": {
        "aliases": {},
        "claims": {},
        "descriptions": {},
        "id": "Q7924",
        "labels": {
            "en": {
                "language": "en",
                "value": "foo"
            }
        },
        "lastrevid": 78256,
        "sitelinks": {},
        "type": "item"
    },
    "success": 1
}
`)
	wikibase := NewClient(client)
	token := "insertokenhere"
	wikibase.editToken = &token
	wikibase.PropertyMap["test"] = "P19"

	item := SingleClaimWithoutInitialUploadTestStruct{Test: "wibble"}
	err := wikibase.CreateItemInstance("blah", &item)

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if item.ID != "Q7924" {
		t.Errorf("ID did not match expected: %s", item)
	}
	if len(item.PropertyIDs) != 0 {
		t.Fatalf("Property map does not contain expected values: %v", item)
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
	if strings.Index(client.MostRecentArgs["data"], "wibble") != -1 {
		t.Errorf("Unexpected data item in API call: %v", client.MostRecentArgs)
	}
}

func TestUploadClaimWihtOmitProperty(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimWithoutInitialUploadTestStruct{Test: "blah"}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(&item, false)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if len(item.PropertyIDs) != 1 {
		t.Fatalf("We expected to have stored a property ID: %v", item)
	}
	if item.PropertyIDs["P14"] != "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63" {
		t.Errorf("We got the wrong property ID: %v", item.PropertyIDs)
	}

	// Check that the request was also sane
	if strings.Index(client.MostRecentArgs["data"], "wibble") != -1 {
		t.Errorf("Unexpected data item in API call: %v", client.MostRecentArgs)
	}
}
