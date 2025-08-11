package main 

import(
	"fmt"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/config"
)

func main(){
	cfg := config.MustLoad()
	fmt.Println(cfg)
}