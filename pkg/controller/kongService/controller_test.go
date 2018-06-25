package kongService

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
		ExpectedUnmarshal *kongClient.Service
		ExpectedErrorNil  bool
	}

	tests := []*TestScenario{
		&TestScenario{
			Name: "ok-basic-1",
			Body: "{\"host\":\"myhost\",\"created_at\":1529400104,\"connect_timeout\":60000,\"id\":\"e5ca22a1-9ff8-425c-b13b-e7a302021775\",\"protocol\":\"http\",\"name\":\"myservice2\",\"read_timeout\":60000,\"port\":3000,\"path\":\"/api\",\"updated_at\":1529400104,\"retries\":10,\"write_timeout\":60000}",
			ExpectedUnmarshal: &kongClient.Service{
				Name:           "myservice2",
				Protocol:       "http",
				Host:           "myhost",
				Port:           3000,
				Path:           "/api",
				Retries:        10,
				ConnectTimeout: 60000,
				WriteTimeout:   60000,
				ReadTimeout:    60000,
				ID:             "e5ca22a1-9ff8-425c-b13b-e7a302021775",
				CreationDate:   1529400104,
				UpdateDate:     1529400104,
			},
			ExpectedErrorNil: true,
		},
		&TestScenario{
			Name:              "ok-fail-json-1",
			Body:              "\"strip_path\"::\"hosts\":[\"example.com\"],\"preserve_host\":true,\"regex_priority\":0,\"updated_at\":1529502375,\"paths\":[\"/\"],\"service\":{\"id\":\"bd3d51da-5c6a-4d9c-b8b9-ca14b30a714e\"},\"methods\":[\"GET\",\"POST\"],\"protocols\":[\"http\"],\"id\":\"6c8dbf33-02f4-4c37-9786-9f41e22f08e7\"}",
			ExpectedUnmarshal: &kongClient.Service{},
			ExpectedErrorNil:  false,
		},
	}

	for _, test := range tests {
		r := http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString(test.Body)),
		}

		res, err := unmarshalService(r.Body)
		errBool := (err == nil)
		if errBool != test.ExpectedErrorNil && !reflect.DeepEqual(res, test.ExpectedUnmarshal) {
			t.Errorf("Test %s failed\nExpected struct %v, got %v\nExpected error == nil %v, got %v", test.Name, test.ExpectedUnmarshal, res, test.ExpectedErrorNil, err)
		}
	}
}
