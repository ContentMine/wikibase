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
	"testing"
)

type SingleClaimTestStruct struct {
	WikiBaseItemHeader

	Test string `property:"test"`
}

func TestUploadClaim(t *testing.T) {

	client := &WikiBaseNetworkTestClient{}
	client.addDataResponse(`
{"pageinfo":{"lastrevid":460},"success":1,"claim":{"mainsnak":{"snaktype":"value","property":"P14","hash":"db735571fef70e4d199d40fe10609312fa8e5fa9","datavalue":{"value":"wot!","type":"string"},"datatype":"string"},"type":"statement","id":"Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63","rank":"normal"}}
`)
	wikibase := NewWikiBaseClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"

	err := wikibase.UploadClaimsForItem(&item)
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
	wikibase := NewWikiBaseClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"
	item.PropertyIDs = make(map[string]string, 0)

	err := wikibase.UploadClaimsForItem(&item)
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
	wikibase := NewWikiBaseClient(client)
	wikibase.PropertyMap["test"] = "P14"
	token := "insertokenhere"
	wikibase.editToken = &token

	item := SingleClaimTestStruct{Test: "blah"}
	item.ID = "Q23"
	item.PropertyIDs = make(map[string]string, 0)
	item.PropertyIDs["P14"] = "Q11$1AE01A5E-EAC8-4568-B866-8E07E93EAB63"

	err := wikibase.UploadClaimsForItem(&item)
	if err != nil {
		t.Fatalf("We got an unexpected error: %v", err)
	}

	if client.InvocationCount != 0 {
		t.Errorf("Got unexpected invocation count: %v", client)
	}
}
