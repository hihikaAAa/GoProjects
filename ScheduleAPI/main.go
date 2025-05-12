package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const apiURL = "https://api.rasp.yandex.net/v3.0"
const apiKey = "0dda023d-6cde-455b-9481-c14d1e75920a"

type StationList struct {
	Countries []Country `json:"countries"`
}
type Country struct {
	Title   string   `json:"title"`
	Regions []Region `json:"regions"`
}
type Region struct {
	Title       string       `json:"title"`
	Settlements []Settlement `json:"settlements"`
}
type Settlement struct {
	Title    string    `json:"title"`
	Codes    CodeBlock `json:"codes"`
	Stations []Station `json:"stations"`
}
type Station struct {
	Title         string    `json:"title"`
	StationType   string    `json:"station_type"`
	TransportType string    `json:"transport_type"`
	Codes         CodeBlock `json:"codes"`
}
type CodeBlock struct {
	YandexCode string `json:"yandex_code"`
}

type SearchResponse struct {
	Segments []RouteSegment `json:"segments"`
}
type RouteSegment struct {
	From StationInfo `json:"from"`
	To           StationInfo `json:"to"`
	Thread       ThreadInfo  `json:"thread"`
	Departure    string      `json:"departure"`
	Arrival      string      `json:"arrival"`
	HasTransfers bool        `json:"has_transfers"`
	Duration     float64     `json:"duration"`
}
type StationInfo struct {
	Title string `json:"title"`
	Code  string `json:"code"`
	Type  string `json:"type"`
}
type ThreadInfo struct {
	TransportType string `json:"transport_type"`
	Title         string `json:"title"`
	Number        string `json:"number"`
}

var stationListCache *StationList

func loadStationList(filename string) (*StationList, error) {
	if stationListCache != nil {
		return stationListCache, nil
	}
	data, err := os.ReadFile(filename)
	if err == nil {
		var list StationList
		if e := json.Unmarshal(data, &list); e == nil {
			stationListCache = &list
			return &list, nil
		}
	}
	url := fmt.Sprintf("%s/stations_list/?apikey=%s&lang=ru_RU&format=json", apiURL, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Ошибка запроса списка станций: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API вернул статус %d при запросе списка станций", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var list StationList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("Не удалось разобрать JSON списка станций: %w", err)
	}
	_ = os.WriteFile(filename, body, 0644)
	stationListCache = &list
	return &list, nil
}

func findSettlementCode(list *StationList, cityName string) (string, error) {
	for _, country := range list.Countries {
		if country.Title == "Россия" {
			for _, region := range country.Regions {
				for _, settlement := range region.Settlements {
					if settlement.Title == cityName {
						return settlement.Codes.YandexCode, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("Город %s не найден в справочнике станций", cityName)
}

func searchRoutes(fromCode, toCode, date string) ([]RouteSegment, error) {
	url := fmt.Sprintf("%s/search/?apikey=%s&format=json&lang=ru_RU&from=%s&to=%s&date=%s&transfers=true",
		apiURL, apiKey, fromCode, toCode, date)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Ошибка HTTP-запроса поиска маршрутов: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API вернул статус %d при поиске маршрутов", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("Не удалось разобрать ответ JSON поиска маршрутов: %w", err)
	}
	return searchResp.Segments, nil
}

func formatRouteSegment(seg RouteSegment) string {
	depTime, _ := time.Parse(time.RFC3339, seg.Departure)
	arrTime, _ := time.Parse(time.RFC3339, seg.Arrival)
	depStr := depTime.Format("15:04")
	arrStr := arrTime.Format("15:04")
	segs := int64(seg.Duration)
	duration := time.Duration(segs) * time.Second
	hrs := int(duration.Hours())
	mins := int(duration.Minutes()) % 60
	durationStr := fmt.Sprintf("%d ч %d мин", hrs, mins)
	transportType := seg.Thread.TransportType
	var transportName string
	switch transportType {
	case "train":
		transportName = "поезд"
	case "suburban":
		transportName = "электричка"
	case "bus":
		transportName = "автобус"
	case "plane":
		transportName = "самолёт"
	default:
		transportName = "транспорт"
	}
	if seg.HasTransfers {
		transferPoint := ""
		if seg.Thread.Title != "" {
			titles := seg.Thread.Title
			firstDash := -1
			lastDash := -1
			for i, ch := range titles {
				if ch == '—' {
					if firstDash == -1 {
						firstDash = i
					}
					lastDash = i
				}
			}
			if firstDash != -1 && lastDash != -1 && firstDash != lastDash {
				transferPoint = titles[firstDash+1 : lastDash]
				transferPoint = strings.TrimSpace(transferPoint)
			}
		}
		if transferPoint == "" {
			transferPoint = "неизвестном пункте"
		}
		return fmt.Sprintf("%s (с пересадкой в %s), выезд в %s, прибытие в %s, в пути %s",
			transportName, transferPoint, depStr, arrStr, durationStr)
	} else {
		return fmt.Sprintf("%s, выезд в %s, прибытие в %s, в пути %s",
			transportName, depStr, arrStr, durationStr)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <YYYY-MM-DD>")
		return
	}
	date := os.Args[1]
	if _, err := time.Parse("2006-01-02", date); err != nil {
		fmt.Printf("Некорректный формат даты: %s. Ожидается YYYY-MM-DD\n", date)
		return
	}

	stationList, err := loadStationList("stations.json")
	if err != nil {
		fmt.Println("Ошибка при загрузке списка станций:", err)
		return
	}

	codeSPB, err1 := findSettlementCode(stationList, "Санкт-Петербург")
	codePskov, err2 := findSettlementCode(stationList, "Псков")
	if err1 != nil || err2 != nil {
		fmt.Println("Не удалось определить код города:", err1, err2)
		return
	}
	routesSPBtoPskov, err := searchRoutes(codeSPB, codePskov, date)
	if err != nil {
		fmt.Println("Ошибка поиска маршрутов СПб -> Псков:", err)
		return
	}
	routesPskovToSPB, err := searchRoutes(codePskov, codeSPB, date)
	if err != nil {
		fmt.Println("Ошибка поиска маршрутов Псков -> СПб:", err)
		return
	}

	fmt.Printf("Маршруты Санкт-Петербург → Псков (%s):\n", date)
	if len(routesSPBtoPskov) == 0 {
		fmt.Println("  Нет доступных рейсов.")
	} else {
		for i, seg := range routesSPBtoPskov {
			fmt.Printf("%d. %s\n", i+1, formatRouteSegment(seg))
		}
	}
	fmt.Printf("\nМаршруты Псков → Санкт-Петербург (%s):\n", date)
	if len(routesPskovToSPB) == 0 {
		fmt.Println("  Нет доступных рейсов.")
	} else {
		for i, seg := range routesPskovToSPB {
			fmt.Printf("%d. %s\n", i+1, formatRouteSegment(seg))
		}
	}
}
