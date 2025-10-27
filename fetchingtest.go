package apiClient2

import (
	apiClient2 "apiClient2/src"
	"fmt"
	"log"
	"net/http"
)

type ApiResponse struct {
	Info    Info        `json:"info"`
	Results []Character `json:"results"`
}

type Info struct {
	Count int     `json:"count"`
	Pages int     `json:"pages"`
	Next  *string `json:"next"` // Puntero para admitir 'null'
	Prev  *string `json:"prev"` // Puntero para admitir 'null'
}

type Character struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Species  string   `json:"species"`
	Type     string   `json:"type"` // 'type' no es palabra clave en Go
	Gender   string   `json:"gender"`
	Origin   Origin   `json:"origin"`
	Location Location `json:"location"`
	Image    string   `json:"image"`
	Episode  []string `json:"episode"`
	URL      string   `json:"url"`
	Created  string   `json:"created"`
}

type Origin struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func main() {

	pool, err := apiClient2.NewPool(5)
	if err != nil {
		log.Fatal(err)
	}

	/*
		character := &ApiResponse{}

		_, err = apiClient2.NewRequest[any, ApiResponse](pool, http.MethodGet, "https://rickandmortyapi.com/api/character", nil, character)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Total characters: %d\n pages %d\n nextPage %d\n prevPage %!   d\n", character.Info.Count, character.Info.Pages, *character.Info.Next, character.Info.Prev)

		for _, char := range character.Results {
			fmt.Println(char.Name)
		} */

	character2 := &ApiResponse{}

	_, err = apiClient2.NewRequest[any, ApiResponse](pool, http.MethodGet, "https://rickandmortyapi.com/api/character?page=2", nil, character2)
	if err != nil {
		log.Fatal(err)
	}

	for _, char := range character2.Results {
		fmt.Printf("character name: %s. character image: %s character gender: %s character origin city: %s \n", char.Name, char.Image, char.Gender, char.Origin.Name)
	}

	fmt.Println(character2.Info.Count, character2.Info.Next)
}
