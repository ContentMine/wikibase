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

type Token struct {
	CSRFToken *string `json:"csrftoken"`
}

type TokensQuery struct {
	Tokens Token `json:"tokens"`
}

type TokenRequestResponse struct {
	GeneralAPIResponse
	Query TokensQuery `json:"query"`
}

type WikiBaseSearchItem struct {
	Duration    int    `json:"ns"`
	Title       string `json:"title"`
	PageID      int    `json:"pageid"`
	DisplayText string `json:"displaytext"`
}

type WikiBaseSearchQuery struct {
	Items []WikiBaseSearchItem `json:"wbsearch"`
}

type WikiBaseSearchResponse struct {
	GeneralAPIResponse
	Query WikiBaseSearchQuery `json:"query"`
}

type WikiBaseArticleEditDetailResponse struct {
	ContentModel  string  `json:"contentmodel"`
	New           *string `json:"new"`
	NewRevisionID int     `json:"newrevid"`
	OldRevisionID int     `json:"oldrevid"`
	NewTimeStamp  string  `json:"newtimestamp"`
	PageID        int     `json:"pageid"`
	Result        string  `json:"result"`
	Title         string  `json:"title"`
}

type WikiBaseError struct {
	Code string `json:"code"`
	Info string `json:"info"`
}

type WikiBaseArticleEditResponse struct {
	Edit  *WikiBaseArticleEditDetailResponse `json:"edit"`
	Error *WikiBaseError                     `json:"error"`
}

type WikiBaseItemLabel struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

type WikiBaseItemEntity struct {
	Labels         map[string]WikiBaseItemLabel `json:"labels"`
	ID             string                       `json:"id"`
	Type           string                       `json:"type"`
	LastRevisionID int                          `json:"lastrevid"`
}

type WikiBaseItemEditResponse struct {
	Entity  *WikiBaseItemEntity `json:"entity"`
	Success int                 `json:"success"`
	Error   *WikiBaseError      `json:"error"`
}
