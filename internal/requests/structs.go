package requests

import (
	"net/http"
	"time"
)

type APIGenesys struct {
	HttpClient   http.Client
	ClientId     string
	ClientSecret string
	AccessToken  string
	UrlBase      string
}

type UserQueryEntities struct {
	Entities  []UserQuery `json:"entities"`
	PageCount int         `json:"pageCount"`
}

type UserQuery struct {
	ID       string `json:"id"`
	Presence struct {
		Source             string `json:"source"`
		PresenceDefinition struct {
			ID             string `json:"id"`
			SystemPresence string `json:"systemPresence"`
			SelfURI        string `json:"selfUri"`
		} `json:"presenceDefinition"`
		ModifiedDate time.Time `json:"modifiedDate"`
	} `json:"presence"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token,omitempty"`
}

type QueryBody struct {
	Interval    string   `json:"interval"`
	UserFilters []Filter `json:"userFilters"`
	Paging      Paging   `json:"paging"`
}

type Filter struct {
	Type       string      `json:"type"`
	Predicates []Predicate `json:"predicates"`
}

type Predicate struct {
	Type      string `json:"type"`
	Dimension string `json:"dimension"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

type Paging struct {
	PageSize   int `json:"pageSize"`
	PageNumber int `json:"pageNumber"`
}

type ResultUserObservation struct {
	UserDetails []UserDetailObservation `json:"userDetails"`
	TotalHits   int                     `json:"totalHits"`
}

type UserDetailObservation struct {
	UserID          string `json:"userId"`
	PrimaryPresence []struct {
		StartTime              time.Time `json:"startTime"`
		EndTime                time.Time `json:"endTime,omitempty"`
		SystemPresence         string    `json:"systemPresence"`
		OrganizationPresenceID string    `json:"organizationPresenceId"`
	} `json:"primaryPresence"`
}
