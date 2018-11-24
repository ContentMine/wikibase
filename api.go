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

// Most API structs are not exported, as they're not exposed by the library API

import (
	"fmt"
)

type WikiBaseType string

const (
	WikiBaseProperty WikiBaseType = "property"
	WikiBaseItem     WikiBaseType = "item"
)

// Error as returned by MediaWiki API
type APIError struct {
	Code string `json:"code"`
	Info string `json:"info"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Error from wikibase %s: %s", e.Code, e.Info)
}

// Mediawiki API response structs

type generalMediaWikiResponse struct {
	BatchComplete string  `json:"batchcomplete"`
	RequestID     *string `json:"requestid"`
}

type editToken struct {
	CSRFToken *string `json:"csrftoken"`
}

type tokensQuery struct {
	Tokens editToken `json:"tokens"`
}

type tokenRequestResponse struct {
	generalMediaWikiResponse
	Query tokensQuery `json:"query"`
}

type searchItem struct {
	Duration    int    `json:"ns"`
	Title       string `json:"title"`
	PageID      int    `json:"pageid"`
	DisplayText string `json:"displaytext"`
}

type searchQuery struct {
	Items []searchItem `json:"wbsearch"`
}

type searchQueryResponse struct {
	generalMediaWikiResponse
	Query searchQuery `json:"query"`
}

type articleEditDetailResponse struct {
	ContentModel  string  `json:"contentmodel"`
	New           *string `json:"new"`
	NewRevisionID int     `json:"newrevid"`
	OldRevisionID int     `json:"oldrevid"`
	NewTimeStamp  string  `json:"newtimestamp"`
	PageID        int     `json:"pageid"`
	Result        string  `json:"result"`
	Title         string  `json:"title"`
}

type articleEditResponse struct {
	Edit  *articleEditDetailResponse `json:"edit"`
	Error *APIError                  `json:"error"`
}

// Wikibase API structs

type itemLabel struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

type itemEntity struct {
	Labels         map[string]itemLabel   `json:"labels"`
	Claims         map[string][]claimInfo `json:"claims"`
	ID             ItemPropertyType       `json:"id"`
	Type           string                 `json:"type"`
	LastRevisionID int                    `json:"lastrevid"`
}

type itemEditResponse struct {
	Entity  *itemEntity `json:"entity"`
	Success int         `json:"success"`
	Error   *APIError   `json:"error"`
}

type pageInfo struct {
	LastRevisionID int `json:"lastrevid"`
}

type snakInfo struct {
	SnakType string `json:"snaktype"`
	Property string `json:"property"`
	Hash     string `json:"hash"`
	DataType string `json:"datatype"`
	// Ignoring datavalue for now...
}

type claimInfo struct {
	MainSnak snakInfo `json:"mainsnak"`
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Rank     string   `json:"rank"`
}

type setCreateResponse struct {
	PageInfo pageInfo  `json:"pageinfo"`
	Success  int       `json:"success"`
	Claim    claimInfo `json:"claim"`
	Error    *APIError `json:"error"`
}

type protection struct {
	Move   *string `json:"move"`
	Edit   *string `json:"edit"`
	Expiry string  `json:"expiry"`
}

type protectDetailResponse struct {
	Title       string       `json:"title"`
	Reason      string       `json:"reason"`
	Protections []protection `json:"protections"`
}

type protectResponse struct {
	Protect *protectDetailResponse `json:"protect"`
	Error   *APIError              `json:"error"`
}
