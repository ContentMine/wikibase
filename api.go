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
)

type WikiBaseType string

const (
	WikiBaseProperty WikiBaseType = "property"
	WikiBaseItem     WikiBaseType = "item"
)

// WikiBase API response structs

type GeneralAPIResponse struct {
	BatchComplete string  `json:"batchcomplete"`
	RequestID     *string `json:"requestid"`
}

type APIToken struct {
	CSRFToken *string `json:"csrftoken"`
}

type TokensQuery struct {
	Tokens APIToken `json:"tokens"`
}

type TokenRequestResponse struct {
	GeneralAPIResponse
	Query TokensQuery `json:"query"`
}

type SearchItem struct {
	Duration    int    `json:"ns"`
	Title       string `json:"title"`
	PageID      int    `json:"pageid"`
	DisplayText string `json:"displaytext"`
}

type SearchQuery struct {
	Items []SearchItem `json:"wbsearch"`
}

type SearchQueryResponse struct {
	GeneralAPIResponse
	Query SearchQuery `json:"query"`
}

type ArticleEditDetailResponse struct {
	ContentModel  string  `json:"contentmodel"`
	New           *string `json:"new"`
	NewRevisionID int     `json:"newrevid"`
	OldRevisionID int     `json:"oldrevid"`
	NewTimeStamp  string  `json:"newtimestamp"`
	PageID        int     `json:"pageid"`
	Result        string  `json:"result"`
	Title         string  `json:"title"`
}

type APIError struct {
	Code string `json:"code"`
	Info string `json:"info"`
}

type ArticleEditResponse struct {
	Edit  *ArticleEditDetailResponse `json:"edit"`
	Error *APIError                  `json:"error"`
}

type ItemLabel struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

type ItemEntity struct {
	Labels         map[string]ItemLabel `json:"labels"`
	ID             ItemPropertyType     `json:"id"`
	Type           string               `json:"type"`
	LastRevisionID int                  `json:"lastrevid"`
}

type ItemEditResponse struct {
	Entity  *ItemEntity `json:"entity"`
	Success int         `json:"success"`
	Error   *APIError   `json:"error"`
}

type PageInfo struct {
	LastRevisionID int `json:"lastrevid"`
}

type SnakInfo struct {
	SnakType string `json:"snaktype"`
	Property string `json:"property"`
	Hash     string `json:"hash"`
	DataType string `json:"datatype"`
	// Ignoring datavalue for now...
}

type ClaimInfo struct {
	MainSnak SnakInfo `json:"mainsnak"`
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Rank     string   `json:"rank"`
}

type ClaimCreateResponse struct {
	PageInfo PageInfo  `json:"pageinfo"`
	Success  int       `json:"success"`
	Claim    ClaimInfo `json:"claim"`
	Error    *APIError `json:"error"`
}

type Protection struct {
	Move   *string `json:"move"`
	Edit   *string `json:"edit"`
	Expiry string  `json:"expiry"`
}

type ProtectDetailResponse struct {
	Title       string       `json:"title"`
	Reason      string       `json:"reason"`
	Protections []Protection `json:"protections"`
}

type ProtectResponse struct {
	Protect *ProtectDetailResponse `json:"protect"`
	Error   *APIError              `json:"error"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Error from wikibase %s: %s", e.Code, e.Info)
}
