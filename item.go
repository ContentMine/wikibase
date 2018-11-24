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
		data.Type = "time"

	case "string":
		t, err := stringClaimToAPIData(value.String())
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, nil
		}
		data.Value = &t
		data.Type = "string"

	case "int":
		t, err := quantityClaimToAPIData(int(value.Int()))
		if err != nil {
			return nil, err
		}
		data.Value = &t
		data.Type = "quantity"

	case "wikibase.ItemPropertyType":
		t, err := itemClaimToAPIData(ItemPropertyType(value.String()))
		if err != nil {
			return nil, err
		}
		data.Value = &t
		data.Type = "wikibase-entityid"

	default:
		return nil, fmt.Errorf("Tried to upload property of unrecognised type %s", full_type_name)
	}

	return &data, nil
}

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
