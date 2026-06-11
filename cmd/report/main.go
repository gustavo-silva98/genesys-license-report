package main

import (
	"genesys-license-report/internal/requests"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	api := requests.APIGenesys{
		HttpClient:   http.Client{},
		ClientId:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		UrlBase:      "https://api.mypurecloud.com",
	}
	api.Authenticate()
	page, err := api.ExtractUsersPage(1)
	if err != nil {
		log.Fatalf("Erro ao extrair pagina 1 %v", err)
		return
	}
	pages, err := api.ExtractUsersPool(page.PageCount, page.PageCount)
	if err != nil {
		log.Fatalf("Erro ao extrair todas as páginas: %v", err)
	}
	userIds := []string{}
	for _, val := range pages {
		if val.Presence.PresenceDefinition.SystemPresence != "" && val.Presence.PresenceDefinition.SystemPresence != "Offline" {
			userIds = append(userIds, val.ID)
		}
	}

}
