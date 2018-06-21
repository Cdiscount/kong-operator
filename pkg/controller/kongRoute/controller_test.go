package controller

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	kongClient "github.com/etiennecoutaud/kong-client-go/kong"
)

func TestUnmarshalRoute(t *testing.T) {

	type TestScenario struct {
		Name              string
		Body              string
		ExpectedUnmarshal *kongClient.Route
		ExpectedError     error
	}

	tests := []*TestScenario{
		&TestScenario{
			Name: "ok-basic-1",
			Body: "{\"created_at\":1529502316,\"strip_path\":true,\"hosts\":[\"example.com\"],\"preserve_host\":true,\"regex_priority\":0,\"updated_at\":1529502375,\"paths\":[\"/\"],\"service\":{\"id\":\"bd3d51da-5c6a-4d9c-b8b9-ca14b30a714e\"},\"methods\":[\"GET\",\"POST\"],\"protocols\":[\"http\"],\"id\":\"6c8dbf33-02f4-4c37-9786-9f41e22f08e7\"}",
			ExpectedUnmarshal: &kongClient.Route{
				Protocols:    []string{"http"},
				Methods:      []string{"GET", "POST"},
				Hosts:        []string{"example.com"},
				Paths:        []string{"/"},
				StripPath:    true,
				PreserveHost: true,
				Service: &kongClient.ServiceRef{
					ID: "bd3d51da-5c6a-4d9c-b8b9-ca14b30a714e",
				},
				ID:           "6c8dbf33-02f4-4c37-9786-9f41e22f08e7",
				CreationDate: 1529502316,
				UpdateDate:   1529502375,
			},
			ExpectedError: nil,
		},
	}

	for _, test := range tests {
		r := http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString(test.Body)),
		}

		res, err := unmarshalRoute(r.Body)
		if !reflect.DeepEqual(res, test.ExpectedUnmarshal) || !reflect.DeepEqual(err, test.ExpectedError) {
			t.Errorf("Test %s failed\nExpected struct %v, got %v\nExpected error %v, got %v", test.Name, test.ExpectedUnmarshal, res, test.ExpectedError, err)
		}
	}
}
