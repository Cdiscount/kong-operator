package kong

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestRouteService_Post(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	input := &Route{Protocols: "http"}

	mux.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		v := new(Route)
		json.NewDecoder(r.Body).Decode(v)
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		testMethod(t, r, "POST")
	})
	_, err := client.Route.Post(input)
	if err != nil {
		t.Errorf("Apis.Post returned error: %v", err)
	}
}
