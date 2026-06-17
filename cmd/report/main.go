package main

import (
	"encoding/csv"
	"fmt"
	"genesys-license-report/internal/requests"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

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
	createReport(api)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		fmt.Println("Ticker disparado at", t)
		createReport(api)
	}

}

func AppendCSVReport(filename string, totalLogged int, toLogout int) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	info, err := file.Stat()
	if info.Size() == 0 {
		writer.Write([]string{"timestamp", "total_logados", "acima_10h"})
	}
	writer.Write([]string{
		time.Now().Format(time.RFC3339),
		strconv.Itoa(totalLogged),
		strconv.Itoa(toLogout),
	})
	return nil
}

func ExportDeslogsSimple(dir string, ids []string) error {
	if dir == "" {
		dir = "exports"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir exports: %w", err)
	}
	name := fmt.Sprintf("deslogs_%s.txt", time.Now().UTC().Format("20060102T150405Z"))
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	for _, id := range ids {
		if _, err := f.WriteString(id + "\n"); err != nil {
			return fmt.Errorf("write id: %w", err)
		}
	}
	return nil
}

func createReport(api requests.APIGenesys) {
	t0 := time.Now()
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
	deslogs := []string{}
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
			log.Printf("TotalHits = 0 %v", val)
			deslogs = append(deslogs, val)
			i++
		}
		//log.Printf("%v - %v\n\n", val, obs)
	}
	for _, val := range observations {
		for _, userDetail := range val.UserDetails {
			// ordena segmentos por startTime asc
			sort.Slice(userDetail.PrimaryPresence, func(i, j int) bool {
				return userDetail.PrimaryPresence[i].StartTime.Before(userDetail.PrimaryPresence[j].StartTime)
			})
			var sessionStart time.Time
			if len(userDetail.PrimaryPresence) == 0 && val.TotalHits > 0 {
				deslogs = append(deslogs, userDetail.UserID)
			}

			// busca o ÚLTIMO segmento OFFLINE de trás pra frente
			foundOffline := false
			for i := len(userDetail.PrimaryPresence) - 1; i >= 0; i-- {
				seg := userDetail.PrimaryPresence[i]
				if seg.SystemPresence == "OFFLINE" {
					sessionStart = seg.EndTime // sessão começou quando o OFFLINE terminou
					foundOffline = true
					break
				}
			}

			// sem OFFLINE na janela = online há mais de 24h
			if !foundOffline && len(userDetail.PrimaryPresence) > 0 {
				sessionStart = userDetail.PrimaryPresence[0].StartTime
				log.Printf("Sem offline - %v\n", userDetail.UserID)
				deslogs = append(deslogs, userDetail.UserID)
			}

			if !sessionStart.IsZero() && now.Sub(sessionStart) > timeTreshold {
				fmt.Printf("%v - sessão desde %v - %v\n", userDetail.UserID, sessionStart, now.Sub(sessionStart))
				deslogs = append(deslogs, userDetail.UserID)
			}

		}

	}
	log.Printf("Observations extraída: %v - Users: %v - Tempo por user: %v", time.Since(t3), len(userIds), time.Since(t3).Seconds()/float64(len(userIds)))
	log.Printf("Tempo total: %v", time.Since(t0))
	log.Println(i)
	log.Println(timeString)
	log.Println(deslogs)
	os.Chdir("..")
	os.Chdir("..")
	AppendCSVReport("reportsHistorical.csv", len(userIds), len(deslogs))
	ExportDeslogsSimple("", deslogs)
	os.Chdir("cmd")
	os.Chdir("report")
}
