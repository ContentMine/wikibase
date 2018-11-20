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
)

type WikiBaseItemHeader struct {
	ID          ItemPropertyType  `json:"wikibase_id,omitempty"`
	PropertyIDs map[string]string `json:"wikibase_property_ids,omitempty"`
}

func (c *WikiBaseClient) UploadClaimsForItem(i interface{}) error {

	// Can we find the headers used to record bits?
	v := reflect.ValueOf(i)
	s := v.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("Expected a struct for item to upload, got %v.", s.Kind())
	}
	header := s.FieldByName("WikiBaseItemHeader")
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

			property_id, ok := c.PropertyMap[tag]
			if ok == false {
				return fmt.Errorf("No property map for property label %s", tag)
			}

			// In future we should make this update the claim, but for now if we've set it once
			// don't set it again
			id_val := property_map_field.MapIndex(reflect.ValueOf(property_id))
			if id_val.IsValid() && id_val.Kind() == reflect.String && len(id_val.String()) > 0 {
				continue
			}

			data, err := getDataForClaim(f, value)
			if err != nil {
				return fmt.Errorf("Failed to marshal %s on %s: %v", property_id, item_id, err)
			}

			id, cerr := c.createClaimOnItem(item_id, property_id, data)
			if cerr != nil {
				return cerr
			}

			property_map_field.SetMapIndex(reflect.ValueOf(property_id), reflect.ValueOf(id))
		}
	}

	return nil
}
