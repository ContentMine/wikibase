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

// Package provides some mediawiki convenience functions along with a pseudo-ORM for creating wikibase items and their
// associated properties.
package wikibase

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// The Wikibase/media wiki client struct. Create this with a call to NewClient, passing it a valid network
// client.
type Client struct {
	client NetworkClientInterface

	// Don't read directly - use GetEditingToken()
	editToken     *string
	editTokenLock sync.RWMutex

	// Mapping of labels to IDs for Items and Properties.
	PropertyMap map[string]string
	ItemMap     map[string]ItemPropertyType
}

func NewClient(oauthClient NetworkClientInterface) *Client {
	return &Client{
		client:      oauthClient,
		PropertyMap: make(map[string]string, 0),
		ItemMap:     make(map[string]ItemPropertyType, 0),
	}
}

// GetEditingToken returns an already acquired editing token for this session, or fetches a new one if necessary. This
// method is thread safe.
func (c *Client) GetEditingToken() (string, error) {

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

	var token tokenRequestResponse
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

func (c *Client) getWikibaseThingIDForLabel(thing WikiBaseType, label string) ([]string, error) {

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
		return nil, err
	}
	defer response.Close()

	var search searchQueryResponse
	err = json.NewDecoder(response).Decode(&search)
	if err != nil {
		return nil, err
	}

	// the search will return close matches not actual matches potentially, so make sure we get exactly
	// matches only
	filtered_items := make([]string, 0)
	for _, item := range search.Query.Items {
		if item.DisplayText == label {

			parts := strings.Split(item.Title, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("We expected type:value in reply, but got: %v", item.Title)
			}
			filtered_items = append(filtered_items, parts[1])
		}
	}

	return filtered_items, nil
}

func (c *Client) FetchPropertyIDsForLabel(label string) ([]string, error) {
	return c.getWikibaseThingIDForLabel(WikiBaseProperty, label)
}

func (c *Client) FetchItemIDsForLabel(label string) ([]string, error) {
	return c.getWikibaseThingIDForLabel(WikiBaseItem, label)
}

// Will create a new mediawiki page if necessary, and set its content to the provided body.
func (c *Client) CreateOrUpdateArticle(title string, body string) (int, error) {

	if len(title) == 0 {
		return 0, fmt.Errorf("Article title must not be an empty string.")
	}

	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return 0, terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action": "edit",
			"token":  editToken,
			"title":  fmt.Sprintf("article:%s", title),
			"text":   body,
		},
	)

	if err != nil {
		return 0, err
	}
	defer response.Close()

	var res articleEditResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return 0, err
	}

	if res.Error != nil {
		return 0, res.Error
	}

	if res.Edit == nil {
		return 0, fmt.Errorf("Unexpected response from server: %v", res)
	}

	return res.Edit.PageID, nil
}

func (c *Client) protectPage(key string, value string) error {

	editToken, terr := c.GetEditingToken()
	if terr != nil {
		return terr
	}

	response, err := c.client.Post(
		map[string]string{
			"action":      "protect",
			"token":       editToken,
			key:           value,
			"protections": "edit=sysop",
			"expiry":      "never",
		},
	)

	if err != nil {
		return err
	}
	defer response.Close()

	var res protectResponse
	err = json.NewDecoder(response).Decode(&res)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return res.Error
	}

	return nil
}

// Given a mediawiki page title set edit protection to sysop for that page. Will fail if page does not exist.
func (c *Client) ProtectPageByTitle(title string) error {
	return c.protectPage("title", title)
}

// Given a mediawiki page ID set edit protection to sysop for that page. Will fail if page does not exist.
func (c *Client) ProtectPageByID(page_id int) error {
	return c.protectPage("pageid", strconv.Itoa(page_id))
}
