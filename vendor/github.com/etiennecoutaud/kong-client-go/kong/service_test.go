package kong

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestService_Post(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	input := &Service{Name: "my-svc"}

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		v := new(Service)
		json.NewDecoder(r.Body).Decode(v)
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		testMethod(t, r, "POST")
	})
	_, err := client.Service.Post(input)
	if err != nil {
		t.Errorf("Apis.Post returned error: %v", err)
	}
}

func TestService_Delete(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	mux.HandleFunc("/services/i", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Service.Delete("i")
	if err != nil {
		t.Errorf("Service.Delete returned error: %v", err)
	}
}
