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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// If you're trying to encode structs to properties then you should use these types
// to help guide the code in encoding things properly for the API
type ItemPropertyType string

// These are the structs to be sent as json in the data section of a wbcreateclaim call. String does not have
// one - the value is direct for string

type itemClaim struct {
	EntityType string `json:"entity-type"`
	NumericID  int    `json:"numeric-id"`
}

type quantityClaim struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

type timeDataClaim struct {
	Time          string `json:"time"`
	TimeZone      int    `json:"timezone"`
	Before        int    `json:"before"`
	After         int    `json:"after"`
	Precision     int    `json:"precision"`
	CalendarModel string `json:"calendarmodel"`
}

// Loading item and property labels from structs

func (c *WikiBaseClient) MapItemConfigurationByLabel(label string) error {
	labels, err := c.FetchItemIDsForLabel(label)
	if err != nil {
		return err
	}
	switch len(labels) {
	case 0:
		return fmt.Errorf("No item ID was found for %s", label)
	case 1:
		c.ItemMap[label] = ItemPropertyType(labels[0])
	default:
		return fmt.Errorf("Multiple item IDs found for %s: %v", labels, labels)
	}
	return nil
}

func (c *WikiBaseClient) MapPropertyAndItemConfiguration(i interface{}) error {

	t := reflect.TypeOf(i)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("property")
		if len(tag) > 0 {
			labels, err := c.FetchPropertyIDsForLabel(tag)
			if err != nil {
				return err
			}
			switch len(labels) {
			case 0:
				return fmt.Errorf("No property ID was found for %s", tag)
			case 1:
				c.PropertyMap[tag] = labels[0]
			default:
				return fmt.Errorf("Multiple property IDs found for %s: %v", tag, labels)
			}
		}

		tag = f.Tag.Get("item")
		if len(tag) > 0 {
			err := c.MapItemConfigurationByLabel(tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Conversation functions

func stringClaimToAPIData(value string) ([]byte, error) {
	return json.Marshal(value)
}

func itemClaimToAPIData(value ItemPropertyType) ([]byte, error) {

	runes := []rune(value)
	if runes[0] != 'Q' {
		return nil, fmt.Errorf("We expected a Q number not %s (starts with %v)", value, runes[0])
	}

	parts := strings.Split(string(value), "Q")
	if len(parts) != 2 {
		return nil, fmt.Errorf("We expected a Q number not %s (splits as %v)", value, parts)
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	item := itemClaim{EntityType: "item", NumericID: id}

	return json.Marshal(item)
}

func quantityClaimToAPIData(value int) ([]byte, error) {

	quantity := quantityClaim{
		Amount: strconv.Itoa(value),
		Unit:   "1",
	}

	return json.Marshal(quantity)
}

func timeDataClaimToAPIData(value string) ([]byte, error) {

	time_data := timeDataClaim{
		Time:          value,
		Precision:     4,
		CalendarModel: "http://www.wikidata.org/entity/Q1985727",
	}

	return json.Marshal(time_data)
}

// Upload properties for structs

func (c *WikiBaseClient) createClaimOnItem(item ItemPropertyType, property_id string, encoded_data []byte) error {

	if len(item) == 0 {
		return fmt.Errorf("Item ID must not be an empty string.")
	}
	if len(property_id) == 0 {
		return fmt.Errorf("Property ID must not be an empty string.")
	}
	if len(encoded_data) == 0 {
		return fmt.Errorf("Encoded data must not be an empty string.")
	}

	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action":   "wbcreateclaim",
			"token":    editToken,
			"entity":   string(item),
			"property": property_id,
			"snaktype": "value",
			"value":    string(encoded_data),
			"bot":      "1",
		},
	)

	if err != nil {
		return err
	}
	defer response.Close()

	var res WikiBaseClaimEditResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return fmt.Errorf("Failed to process claim %s on %s with data %v: %v", property_id, item,
			string(encoded_data), res.Error)
	}

	if res.Success != 1 {
		return fmt.Errorf("We got an unexpected success value adding claim %s on %s with data %v: %v", property_id,
			item, string(encoded_data), res)
	}

	return nil

}

func getDataForClaim(f reflect.StructField, value reflect.Value) ([]byte, error) {

	// now work out how to encode this. We currently support: string, int (as quantity), Time (as TimeData),
	// and ItemPropertyType (as an item). Everything else we just raise an error on.

	var data []byte

	full_type_name := fmt.Sprintf("%v", f.Type)
	switch full_type_name {
	case "time.Time":
		m, ok := value.Interface().(json.Marshaler)
		if !ok {
			return nil, fmt.Errorf("time.Time does not respect JSON marshalling any more.")
		}
		var err error
		data, err = m.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return timeDataClaimToAPIData(string(data))
	case "string":
		return stringClaimToAPIData(value.String())
	case "int":
		return quantityClaimToAPIData(int(value.Int()))
	case "wikibase.ItemPropertyType":
		return itemClaimToAPIData(ItemPropertyType(value.String()))
	default:
		return nil, fmt.Errorf("Tried to upload property of unrecognised type %s", full_type_name)
	}
}

func (c *WikiBaseClient) UploadClaimsForItem(item ItemPropertyType, i interface{}) error {

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		value := v.Field(i)

		tag := f.Tag.Get("property")
		if len(tag) > 0 {

			property_id, ok := c.PropertyMap[tag]
			if ok == false {
				return fmt.Errorf("No property map for property label %s", tag)
			}

			data, err := getDataForClaim(f, value)
			if err != nil {
				return err
			}

			err = c.createClaimOnItem(item, property_id, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
