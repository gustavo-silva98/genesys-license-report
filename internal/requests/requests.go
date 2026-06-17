package requests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func (api *APIGenesys) ExtractUsersPool(numPages int, numWorkers int) ([]UserQuery, error) {
	jobs := make(chan int, numPages)
	results := make(chan []UserQuery, numPages)
	totalPages := make([]UserQuery, numPages*500)

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go api.WorkerExtractUser(&wg, jobs, results)
	}

	for i := 1; i <= numPages; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(results)

	for elem := range results {
		for _, val := range elem {
			totalPages = append(totalPages, val)
		}
	}
	return totalPages, nil
}

func (api *APIGenesys) ExtractUsersPage(pageNumber int) (UserQueryEntities, error) {

	var response UserQueryEntities
	log.Printf("Extraindo pagina Users %v\n", pageNumber)
	url := fmt.Sprintf("%v/api/v2/users?pageSize=500&pageNumber=%v&expand=presence&state=active", api.UrlBase, pageNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return UserQueryEntities{}, nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := api.HttpClient.Do(req)
	if err != nil {
		return UserQueryEntities{}, nil
	}
	if res.StatusCode != 200 {
		msg := fmt.Sprintf("Falha ao extrair usuários - HTTP %v", res.StatusCode)
		return UserQueryEntities{}, errors.New(msg)
	}

	body, _ := io.ReadAll(res.Body)

	json.Unmarshal(body, &response)
	return response, nil
}

func (api *APIGenesys) Authenticate() AuthResponse {
	var response AuthResponse

	authURL, _ := url.Parse("https://login.mypurecloud.com/oauth/token")

	request := http.Request{
		URL:    authURL,
		Close:  true,
		Method: "POST",
		Header: make(map[string][]string),
	}

	// Seta Header com basic Token (token autenticado)
	request.Header.Set("Authorization", fmt.Sprintf(
		"Basic %v", base64.StdEncoding.EncodeToString([]byte(api.ClientId+":"+api.ClientSecret)),
	),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Seta o Body com client_credentials
	// TODO validar url.Values
	formParams := url.Values{}
	formParams["grant_type"] = []string{"client_credentials"}
	request.Body = io.NopCloser(strings.NewReader(formParams.Encode()))

	// Faz a request de fato
	res, err := api.HttpClient.Do(&request)
	// TODO Fazer o tratamento melhor de erros?
	if err != nil {
		log.Println(err)
		panic(err)
	}

	// lê o Body da request
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		log.Println(res.StatusCode)
		panic(res.Status)
	}

	// Transpassa o body pra struct AuthResponse
	json.Unmarshal(body, &response)

	api.AccessToken = response.AccessToken
	return response
}

func (api *APIGenesys) WorkerExtractUser(wg *sync.WaitGroup, jobs <-chan int, results chan<- []UserQuery) {
	defer wg.Done()

	for val := range jobs {
		page, err := api.ExtractUsersPage(val)
		if err != nil {
			log.Printf("erro ao extrair página de usuários %v: %v", val, err)
			return
		}
		results <- page.Entities
	}
}

func (api *APIGenesys) GetUserObservation(body []byte) (ResultUserObservation, error) {
	var result ResultUserObservation
	url := fmt.Sprintf("%v/api/v2/analytics/users/details/query", api.UrlBase)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("erro ao montar Request: %v", err)
		return ResultUserObservation{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api.AccessToken)

	resp, err := api.HttpClient.Do(req)
	if err != nil {
		log.Printf("erro ao fazer request: %v", err)
		return ResultUserObservation{}, err
	}
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("erro ao ler body de resultado userObservation: %v", err)
		return ResultUserObservation{}, err
	}
	defer resp.Body.Close()

	json.Unmarshal(respBody, &result)
	return result, nil
}

func BuildUserQueryBody(userId string, interval string, pageSize int, pageNumber int) ([]byte, error) {
	predicate := Predicate{
		Type:      "dimension",
		Dimension: "userId",
		Operator:  "matches",
		Value:     userId,
	}
	body := QueryBody{
		Interval: interval,
		UserFilters: []Filter{
			{Type: "and", Predicates: []Predicate{predicate}},
		},
		Paging: Paging{PageSize: pageSize, PageNumber: pageNumber},
	}
	return json.Marshal(body)
}
