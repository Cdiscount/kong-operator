package kong

import (
	"fmt"
	"net/http"
)

type ServiceService service

type Service struct {
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	Retries        int    `json:"retries"`
	ConnectTimeout int    `json:"connect_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	ReadTimeout    int    `json:"read_timeout"`
	URL            string `json:"url,omitempty"`
	ID             string `json:"id,omitempty"`
	CreationDate   int    `json:"created_at,omitempty"`
	UpdateDate     int    `json:"updated_at,omitempty"`
}

//Post Add new Kong Service /services
func (ss *ServiceService) Post(s *Service) (*http.Response, error) {
	req, err := ss.client.NewRequest("POST", "services", s)
	if err != nil {
		return nil, err
	}
	resp, err := ss.client.Do(req, nil)
	return resp, err
}

//Delete Delete Kong Service based on ID
func (ss *ServiceService) Delete(serviceID string) (*http.Response, error) {
	u := fmt.Sprintf("services/%s", serviceID)

	req, err := ss.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ss.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

//Patch Update Kong Service based on ID
func (ss *ServiceService) Patch(serviceID string, s *Service) (*http.Response, error) {
	u := fmt.Sprintf("services/%s", serviceID)

	req, err := ss.client.NewRequest("PATCH", u, s)
	if err != nil {
		return nil, err
	}

	resp, err := ss.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}
