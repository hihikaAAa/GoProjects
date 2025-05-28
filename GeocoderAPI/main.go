package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const apiURL = "https://geocode-maps.yandex.ru/1.x/"

const apiKey = "fb6987dc-1256-4a15-953a-75f3fe0c16c3"

type Coords struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type Cache map[string]Coords

type yResp struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					Point struct {
						Pos string `json:"pos"`
					} `json:"Point"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

func loadCache(path string) (Cache, error) {
	if b, err := os.ReadFile(path); err == nil {
		var c Cache
		if json.Unmarshal(b, &c) == nil {
			return c, nil
		}
	}
	return make(Cache), nil
}

func saveCache(path string, c Cache) {
	if b, err := json.MarshalIndent(c, "", "  "); err == nil {
		_ = os.WriteFile(path, b, 0644)
	}
}

func geocode(addr string, c Cache) (Coords, error) {
	if v, ok := c[addr]; ok {
		return v, nil
	}
	q := fmt.Sprintf("%s?apikey=%s&format=json&geocode=%s", apiURL, apiKey, url.QueryEscape(addr))
	resp, err := http.Get(q)
	if err != nil {
		return Coords{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r yResp
	if err := json.Unmarshal(body, &r); err != nil {
		return Coords{}, err
	}
	if len(r.Response.GeoObjectCollection.FeatureMember) == 0 {
		return Coords{}, fmt.Errorf("не найдено")
	}
	pos := r.Response.GeoObjectCollection.FeatureMember[0].GeoObject.Point.Pos
	parts := strings.Split(pos, " ")
	if len(parts) != 2 {
		return Coords{}, fmt.Errorf("неверный формат pos")
	}
	coord := Coords{Lat: parts[1], Lon: parts[0]}
	c[addr] = coord
	saveCache("cache.json", c)
	return coord, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <адрес>")
		return
	}
	addr := strings.Join(os.Args[1:], " ")
	cache, _ := loadCache("cache.json")
	coord, err := geocode(addr, cache)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Координаты «%s»: %s, %s\n", addr, coord.Lat, coord.Lon)
}
