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
	"testing"
	"time"
    "unicode/utf8"
)


func TestStringClaimEncode(t *testing.T) {

    const testdata = "hello, world"

    data, err := stringClaimToAPIData(testdata)
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }
    // A cheap test, but check that the returned string is two characters longer due to being in quotes
    if utf8.RuneCountInString(testdata) + 2 != utf8.RuneCountInString(string(data)) {
        t.Fatalf("Length of encoded data not what we expected: %v", string(data))
    }
}

func TestItemClaimEncode(t *testing.T) {
    _, err := itemClaimToAPIData("Q42")
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }
}

func TestPropertyAsItemClaimEncode(t *testing.T) {
    _, err := itemClaimToAPIData("P42")
    if err == nil {
        t.Fatalf("We got an expected an error")
    }
}

func TestInvalidItemClaimEncode(t *testing.T) {
    _, err := itemClaimToAPIData("42")
    if err == nil {
        t.Fatalf("We got an expected an error")
    }
}

func TestQuntityClaimEncode(t *testing.T) {
    _, err := quantityClaimToAPIData(42)
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }
}

func TestTimeDataClaimEncode(t *testing.T) {
    _, err := timeDataClaimToAPIData(time.Now())
    if err != nil {
        t.Fatalf("We got an unexpected error: %v", err)
    }
}
