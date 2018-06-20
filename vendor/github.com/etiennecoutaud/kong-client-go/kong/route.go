package kong

import (
	"fmt"
	"net/http"
)

type RouteService service

type ServiceRef struct {
	ID string `json:"id,omitempty"`
}

type Route struct {
	Protocols    []string    `json:"protocols,omitempty"`
	Methods      []string    `json:"methods,omitempty"`
	Hosts        []string    `json:"hosts,omitempty"`
	Paths        []string    `json:"paths,omitempty"`
	StripPath    bool        `json:"strip_path,omitempty"`
	PreserveHost bool        `json:"preserve_host,omitempty"`
	Service      *ServiceRef `json:"service,omitempty"`
	ID           string      `json:"id,omitempty"`
	CreationDate int         `json:"created_at,omitempty"`
	UpdateDate   int         `json:"updated_at,omitempty"`
}

//Post Create new Kong Route API Object
func (s *RouteService) Post(route *Route) (*http.Response, error) {
	req, err := s.client.NewRequest("POST", "routes", route)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)

	return resp, err
}

//Delete Delete Kong Route API Object
func (s *RouteService) Delete(routeID string) (*http.Response, error) {
	u := fmt.Sprintf("routes/%v", routeID)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

//Patch Update Kong Service based on ID
func (s *RouteService) Patch(routeID string, r *Route) (*http.Response, error) {
	u := fmt.Sprintf("routes/%s", routeID)

	req, err := s.client.NewRequest("PATCH", u, r)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}
