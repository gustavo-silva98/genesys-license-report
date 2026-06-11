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
