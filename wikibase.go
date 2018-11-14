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
	"strings"
	"sync"
)

type WikiBaseClient struct {
	client WikiBaseOAuthClientInterface

	// Don't read directly - use GetEditingToken()
	editToken     *string
	editTokenLock sync.RWMutex
}

func NewWikiBaseClient(oauthClient WikiBaseOAuthClientInterface) *WikiBaseClient {
	return &WikiBaseClient{client: oauthClient}
}

func (c *WikiBaseClient) GetEditingToken() (string, error) {

	c.editTokenLock.RLock()
	initVal := c.editToken
	c.editTokenLock.RUnlock()

	if initVal != nil {
		return *initVal, nil
	}

	c.editTokenLock.Lock()
	defer c.editTokenLock.Unlock()

	// at start of day there's a big risk all go-routines race on getting
	// the edit token, so bail early if someone else has won
	if c.editToken != nil {
		return *c.editToken, nil
	}

	response, err := c.client.Get(
		map[string]string{
			"action": "query",
			"meta":   "tokens",
		},
	)

	if err != nil {
		return "", err
	}
	defer response.Close()

	var token TokenRequestResponse
	err = json.NewDecoder(response).Decode(&token)
	if err != nil {
		return "", err
	}

	if token.Query.Tokens.CSRFToken == nil {
		return "", fmt.Errorf("Failed to get token in response from server: %v", token)
	}

	c.editToken = token.Query.Tokens.CSRFToken

	return *c.editToken, nil
}

func (c *WikiBaseClient) getWikibaseThingForLabel(thing WikiBaseType, label string) (string, error) {

	response, err := c.client.Get(
		map[string]string{
			"action":      "query",
			"list":        "wbsearch",
			"wbssearch":   label,
			"wbstype":     string(thing),
			"wbslanguage": "en",
		},
	)

	if err != nil {
		return "", err
	}
	defer response.Close()

	var search WikiBaseSearchResponse
	err = json.NewDecoder(response).Decode(&search)
	if err != nil {
		return "", err
	}

	switch len(search.Query.Items) {
	case 0:
		return "", fmt.Errorf("Failed to find any matching properties for %s", label)
	case 1:
		parts := strings.Split(search.Query.Items[0].Title, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("We expected type:value in reply, but got: %v", search.Query.Items[0].Title)
		}
		return parts[1], nil // TODO fix
	default:
		return "", fmt.Errorf("Too many matches returned for %s: %d", label, len(search.Query.Items))
	}
}

func (c *WikiBaseClient) GetPropertyForLabel(label string) (string, error) {
	return c.getWikibaseThingForLabel(WikiBaseProperty, label)
}

func (c *WikiBaseClient) GetItemForLabel(label string) (string, error) {
	return c.getWikibaseThingForLabel(WikiBaseItem, label)
}

func (c *WikiBaseClient) CreateArticle(title string, body string) (int, error) {

	if len(title) == 0 {
		return 0, fmt.Errorf("Article title must not be an empty string.")
	}

	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return 0, terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action":     "edit",
			"token":      editToken,
			"createonly": "true",
			"title":      fmt.Sprintf("article:%s", title),
			"text":       body,
		},
	)

	if err != nil {
		return 0, err
	}
	defer response.Close()

	var res WikiBaseArticleEditResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return 0, err
	}

	if res.Error != nil {
		return 0, fmt.Errorf("%s: %s", res.Error.Code, res.Error.Info)
	}

	if res.Edit == nil {
		return 0, fmt.Errorf("Unexpected response from server: %v", res)
	}

	return res.Edit.PageID, nil
}

func (c *WikiBaseClient) CreateItemInstance(label string) (string, error) {

	if len(label) == 0 {
		return "", fmt.Errorf("Item label must not be an empty string.")
	}

	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return "", terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action": "wbeditentity",
			"token":  editToken,
			"new":    "item",
			"data":   fmt.Sprintf("{\"labels\": {\"en\": {\"language\": \"en\", \"value\": \"%s\"}}}", label),
		},
	)

	if err != nil {
		return "", err
	}
	defer response.Close()

	var res WikiBaseItemEditResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return "", err
	}

	if res.Error != nil {
		return "", fmt.Errorf("%s: %s", res.Error.Code, res.Error.Info)
	}

	if res.Success != 1 {
		return "", fmt.Errorf("We got an unexpected success value: %v", res)
	}

	if res.Entity == nil {
		return "", fmt.Errorf("Unexpected response from server: %v", res)
	}

	return res.Entity.ID, nil
}
