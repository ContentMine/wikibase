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
	"io"
	"os"

	"github.com/mrjones/oauth"
)

// We don't use the OAuth interface directly so as to let us more readily write unit tests and save on boilerplate
// code

type NetworkClientInterface interface {
	Get(args map[string]string) (io.ReadCloser, error)
	Post(args map[string]string) (io.ReadCloser, error)
}

// Structured used to hold the consumer and access tokens, such that they can be serialised readily

type ConsumerInformation struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type AccessToken struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type OAuthInformation struct {
	Consumer ConsumerInformation `json:"consumer"`
	Access   *AccessToken        `json:"access,omitempty"`
}

type OAuthNetworkClient struct {
	APIURL string

	AccessToken *oauth.AccessToken
	consumer    *oauth.Consumer
}

// Factory method for creating a new client

func LoadOauthInformation(path string) (OAuthInformation, error) {
	var info OAuthInformation

	f, err := os.Open(path)
	if err != nil {
		return OAuthInformation{}, err
	}

	err = json.NewDecoder(f).Decode(&info)
	return info, err
}

func NewOAuthNetworkClient(oauthInfo OAuthInformation, urlbase string) *OAuthNetworkClient {

	res := OAuthNetworkClient{
		APIURL: fmt.Sprintf("%s/w/api.php", urlbase),
	}

	if oauthInfo.Access != nil {
		aToken := oauth.AccessToken{Token: oauthInfo.Access.Token, Secret: oauthInfo.Access.Secret}
		res.AccessToken = &aToken
	}

	res.consumer = oauth.NewConsumer(
		oauthInfo.Consumer.Key,
		oauthInfo.Consumer.Secret,
		oauth.ServiceProvider{
			RequestTokenUrl:   fmt.Sprintf("%s/wiki/Special:OAuth/initiate", urlbase),
			AuthorizeTokenUrl: fmt.Sprintf("%s/wiki/Special:OAuth/authorize", urlbase),
			AccessTokenUrl:    fmt.Sprintf("%s/wiki/Special:OAuth/token", urlbase),
		})

	return &res
}

// Network action requests
//
// These methods should do as little as possible beyond abstracting the network protocol to enable us
// to do testing. This is why they don't do JSON demarshalling here, as that needs to be tested.

func (client *OAuthNetworkClient) Get(args map[string]string) (io.ReadCloser, error) {

	// We always deal in JSON here
	args["format"] = "json"

	response, err := client.consumer.Get(client.APIURL, args, client.AccessToken)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Go a %d response: %s", response.StatusCode, response.Status)
	}

	return response.Body, nil
}

func (client *OAuthNetworkClient) Post(args map[string]string) (io.ReadCloser, error) {

	// We always deal in JSON here
	args["format"] = "json"

	response, err := client.consumer.Post(client.APIURL, args, client.AccessToken)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Go a %d response: %s", response.StatusCode, response.Status)
	}

	return response.Body, nil
}
