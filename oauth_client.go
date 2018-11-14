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
	"io"

	"github.com/mrjones/oauth"
)

// We don't use the OAuth interface directly so as to let us more readily write unit tests and save on boilerplate
// code

type WikiBaseOAuthClientInterface interface {
	Get(args map[string]string) (io.ReadCloser, error)
	Post(args map[string]string) (io.ReadCloser, error)
}

type WikiBaseOAuthClient struct {
	APIURL string

	AccessToken *oauth.AccessToken
	consumer    *oauth.Consumer
}

// Factory method for creating a new client
func NewOAuthClient(consumerKey string, consumerSecret string, urlbase string, accessToken *oauth.AccessToken) *WikiBaseOAuthClient {

	res := WikiBaseOAuthClient{
		APIURL:      fmt.Sprintf("%s/w/api.php", urlbase),
		AccessToken: accessToken,
	}

	res.consumer = oauth.NewConsumer(
		consumerKey,
		consumerSecret,
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

func (client *WikiBaseOAuthClient) Get(args map[string]string) (io.ReadCloser, error) {

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

func (client *WikiBaseOAuthClient) Post(args map[string]string) (io.ReadCloser, error) {

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
