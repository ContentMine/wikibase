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
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ItemHeader must be embedded in all structs that are to be uploaded to Wikibase. If you give this embedded struct
// a JSON Tag then you can save and restore the Wikibase ID state for the entire struct, which can be used to avoid
// creating the item multiple times when you run.
//
// Each field that you want to sync as a property on the item in Wikibase must have a "property" tag, with the name
// of the label of a property on Wikibase. We use labels rather than P numbers as it's impossible to guarantee that
// production, staging, and test servers will use the same P numbers as they are allocated automatically by the
// Wikibase server; labels on the other hand can be managed by humans/bots. You should always call the client function
// MapPropertyAndItemConfiguration to populate it's internal map before attempting to create/update Items and their
// properties. If you add an "omitoncreate" clause then the Property will not be added to the item at create time,
// only later on during property sync.
type ItemHeader struct {
	ID          ItemPropertyType  `json:"wikibase_id,omitempty"`
	PropertyIDs map[string]string `json:"wikibase_property_ids,omitempty"`
}

type dataValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type snakCreateInfo struct {
	DataValue *dataValue `json:"datavalue"`
	Property  string     `json:"property"`
	SnakType  string     `json:"snaktype"`
}

type claimCreate struct {
	MainSnak snakCreateInfo `json:"mainsnak"`
	Rank     string         `json:"rank"`
	Type     string         `json:"type"`
}

type itemCreateData struct {
	Labels map[string]itemLabel `json:"labels"`
	Claims []claimCreate        `json:"claims"`
}

func getItemCreateClaimValue(f reflect.StructField, value reflect.Value) (*dataValue, error) {

	full_type_name := fmt.Sprintf("%v", f.Type)

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil, nil
		} else {
			value = value.Elem()
			if full_type_name[0] != '*' {
				return nil, fmt.Errorf("We expected a pointer type: %s", full_type_name)
			}
			full_type_name = full_type_name[1:]
		}
	}

	data := dataValue{}

	datatype, err := goTypeToWikibaseType(f)
	if err != nil {
		return nil, err
	}

	switch full_type_name {
	case "time.Time":
		m, ok := value.Interface().(encoding.TextMarshaler)
		if !ok {
			return nil, fmt.Errorf("time.Time does not respect JSON marshalling any more.")
		}
		b, berr := m.MarshalText()
		if berr != nil {
			return nil, berr
		}
		t, err := timeDataClaimToAPIData(string(b))
		if err != nil {
			return nil, err
		}
		data.Value = &t
		data.Type = datatype

	case "string":
		t, err := stringClaimToAPIData(value.String())
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, nil
		}
		data.Value = &t
		data.Type = datatype

	case "int":
		t, err := quantityClaimToAPIData(int(value.Int()))
		if err != nil {
			return nil, err
		}
		data.Value = &t
		data.Type = datatype

	case "wikibase.ItemPropertyType":
		t, err := itemClaimToAPIData(ItemPropertyType(value.String()))
		if err != nil {
			return nil, err
		}
		data.Value = &t
		data.Type = datatype

	default:
		return nil, fmt.Errorf("Tried to upload property of unrecognised type %s", full_type_name)
	}

	return &data, nil
}

// CreateItemInstance will take a pointer to a Go structure that has the embedded wikibase header and
// item and property tags on its fields and create a new item with the provided label. Any fields in the structure
// with a Property tag that does not contain the "omitoncreate" clause will also be created as item claims at the
// same time.
func (c *Client) CreateItemInstance(label string, i interface{}) error {

	if len(label) == 0 {
		return fmt.Errorf("Item label must not be an empty string.")
	}

	// Can we find the headers used to record bits?
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected a pointer to the item to upload, not %v", v.Kind())
	}
	s := v.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("Expected a struct for item to upload, got %v.", s.Kind())
	}
	header := s.FieldByName("ItemHeader")
	if !header.IsValid() {
		return fmt.Errorf("Expected struct to have item header")
	}

	// Are there any properties that we should create at this venture as part of initial
	// upload?
	claims := make([]claimCreate, 0)

	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		value := s.Field(i)

		tag := f.Tag.Get("property")
		if len(tag) > 0 {
			// There may be multiple tags, the first one of which is the property name
			parts := strings.Split(tag, ",")
			tag = parts[0]

			// if there's a omitoncreate then skip this
			skiptag := false
			for _, t := range parts[1:] {
				if t == "omitoncreate" {
					skiptag = true
					break
				}
			}
			if skiptag {
				continue
			}

			property_id, ok := c.PropertyMap[tag]
			if ok == false {
				return fmt.Errorf("No property map for property label %s", tag)
			}

			claim, err := getItemCreateClaimValue(f, value)
			if err != nil {
				return fmt.Errorf("Failed to marshal %s during create: %v", property_id, err)
			}

			snaktype := "value"
			if claim == nil {
				snaktype = "novalue"
			}
			create := claimCreate{
				MainSnak: snakCreateInfo{
					DataValue: claim,
					Property:  property_id,
					SnakType:  snaktype,
				},
				Rank: "normal",
				Type: "statement",
			}

			claims = append(claims, create)
		}
	}

	labels := make(map[string]itemLabel, 0)
	labels["en"] = itemLabel{Language: "en", Value: label}
	item := itemCreateData{Labels: labels, Claims: claims}

	b, berr := json.Marshal(&item)
	if berr != nil {
		return berr
	}

	// Having got things
	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action": "wbeditentity",
			"token":  editToken,
			"new":    "item",
			"data":   string(b),
		},
	)

	if err != nil {
		return err
	}
	defer response.Close()

	var res itemEditResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return res.Error
	}

	if res.Success != 1 {
		return fmt.Errorf("We got an unexpected success value: %v", res)
	}

	if res.Entity == nil {
		return fmt.Errorf("Unexpected response from server: %v", res)
	}

	// We now need to extract the ID and all the property IDs we created
	id_field := header.FieldByName("ID")
	if !id_field.IsValid() || id_field.Kind() != reflect.String {
		return fmt.Errorf("Expected header to have string ID field")
	}
	if !id_field.CanSet() {
		return fmt.Errorf("Expected item header to be mutable!")
	}
	id_field.SetString(string(res.Entity.ID))

	// we need the map used to store property IDs
	property_map_field := header.FieldByName("PropertyIDs")
	if !property_map_field.IsValid() || property_map_field.Kind() != reflect.Map {
		return fmt.Errorf("Expected header to have a property map")
	}
	if property_map_field.IsNil() {
		property_map_field.Set(reflect.MakeMap(property_map_field.Type()))
	}

	for property, claims := range res.Entity.Claims {
		// In theory there can be multiple claims per property, but we only support creating one at the moment
		// so error if there's more than one
		if len(claims) > 1 {
			return fmt.Errorf("Unexpected list of claims for %s after we created just one: %v", property, claims)
		} else if len(claims) == 1 {
			property_map_field.SetMapIndex(reflect.ValueOf(property), reflect.ValueOf(claims[0].ID))
		}
	}

	return nil
}

// UploadClaimsForItem will take a pointer to a Go structure that has the embedded wikibase header and
// item and property tags on its fields and set the claims on the item to match. The item must have been created
// already. If allow_refresh is set to true, all properties will be written, regardless of whether they've been
// uploaded before; if set to false only items with no existing Wikibase Property ID in the map will be updated.
func (c *Client) UploadClaimsForItem(i interface{}, allow_refresh bool) error {

	// Can we find the headers used to record bits?
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected a pointer to the item to upload, not %v", v.Kind())
	}
	s := v.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("Expected a struct for item to upload, got %v.", s.Kind())
	}
	header := s.FieldByName("ItemHeader")
	if !header.IsValid() {
		return fmt.Errorf("Expected struct to have item header")
	}

	// Having got the header, get the item ID
	id_field := header.FieldByName("ID")
	if !id_field.IsValid() || id_field.Kind() != reflect.String {
		return fmt.Errorf("Expected header to have string ID field")
	}
	item_id := ItemPropertyType(id_field.String())
	if len(item_id) == 0 {
		return fmt.Errorf("Item ID is nil in item")
	}

	// we need the map used to store property IDs
	property_map_field := header.FieldByName("PropertyIDs")
	if !property_map_field.IsValid() || property_map_field.Kind() != reflect.Map {
		return fmt.Errorf("Expected header to have a property map")
	}
	if property_map_field.IsNil() {
		property_map_field.Set(reflect.MakeMap(property_map_field.Type()))
	}

	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		value := s.Field(i)

		tag := f.Tag.Get("property")
		if len(tag) > 0 {

			// There may be multiple tags, the first one of which is the property name
			parts := strings.Split(tag, ",")
			tag = parts[0]

			property_id, ok := c.PropertyMap[tag]
			if ok == false {
				return fmt.Errorf("No property map for property label %s", tag)
			}

			// In future we should make this update the claim, but for now if we've set it once
			// don't set it again
			id_val := property_map_field.MapIndex(reflect.ValueOf(property_id))
			have_existing_claim := false
			if id_val.IsValid() && id_val.Kind() == reflect.String && len(id_val.String()) > 0 {
				have_existing_claim = true
			}

			data, err := getDataForClaim(f, value)
			if err != nil {
				return fmt.Errorf("Failed to marshal %s on %s: %v", property_id, item_id, err)
			}

			if !have_existing_claim {
				id, err := c.createClaimOnItem(item_id, property_id, data)
				if err != nil {
					return err
				}

				property_map_field.SetMapIndex(reflect.ValueOf(property_id), reflect.ValueOf(id))
			} else if allow_refresh {
				err := c.updateClaim(id_val.String(), data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
