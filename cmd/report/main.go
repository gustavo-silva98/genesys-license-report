package main

import (
	"fmt"
	"genesys-license-report/internal/requests"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	t0 := time.Now()
	godotenv.Load()
	api := requests.APIGenesys{
		HttpClient:   http.Client{},
		ClientId:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		UrlBase:      "https://api.mypurecloud.com",
	}
	api.Authenticate()

	log.Printf("Autenticação feita: %v\n", time.Since(t0))
	t1 := time.Now()
	page, err := api.ExtractUsersPage(1)
	if err != nil {
		log.Fatalf("Erro ao extrair pagina 1 %v", err)
		return
	}
	pages, err := api.ExtractUsersPool(page.PageCount, page.PageCount)
	if err != nil {
		log.Fatalf("Erro ao extrair todas as páginas: %v", err)
	}
	log.Printf("Usuários extraídos: %v\n", time.Since(t1))
	t2 := time.Now()
	userIds := []string{}
	for _, val := range pages {
		if val.Presence.PresenceDefinition.SystemPresence != "" && val.Presence.PresenceDefinition.SystemPresence != "Offline" {
			userIds = append(userIds, val.ID)
		}
	}
	log.Printf("Separação de usuários feita: %v\n", time.Since(t2))
	//fmt.Println(userIds)
	fmt.Println(len(userIds))
	t3 := time.Now()
	now := time.Now()
	timeTreshold := 10*time.Hour + 30*time.Minute
	timeString := fmt.Sprintf("%v/%v", time.Now().Add(-24*time.Hour).Format(time.RFC3339Nano), time.Now().Format(time.RFC3339Nano))
	observations := []requests.ResultUserObservation{}
	var i int
	for _, val := range userIds {
		body, err := requests.BuildUserQueryBody(val, timeString, 100, 1)
		if err != nil {
			log.Println(err)
			return
		}
		obs, err := api.GetUserObservation(body)
		if err != nil {
			log.Println(err)
			return
		}
		observations = append(observations, obs)
		if obs.TotalHits == 0 {
			i++
		}
		//log.Printf("%v - %v\n\n", val, obs)
	}
	deslogs := 0
	for _, val := range observations {
		for _, userDetail := range val.UserDetails {
			for _, presence := range userDetail.PrimaryPresence {
				if presence.SystemPresence == "OFFLINE" {
					diff := now.Sub(presence.EndTime)
					if diff > timeTreshold {
						fmt.Printf("%v - %v - %v\n\n", userDetail.UserID, presence, diff)
						deslogs++
						continue
					}
				}
			}
		}
	}

	log.Printf("Observations extraída: %v - Users: %v - Tempo por user: %v", time.Since(t3), len(userIds), time.Since(t3).Seconds()/float64(len(userIds)))
	log.Printf("Tempo total: %v", time.Since(t0))
	log.Println(timeString)
	log.Println(i)
	log.Println(deslogs)
}
