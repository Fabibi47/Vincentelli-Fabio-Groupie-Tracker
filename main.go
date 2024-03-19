package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
)

type User struct {
	UserName        string   `json:"username"`
	Password        string   `json:"password"`
	FavoriteTracks  []string `json:"favorite_tracks"`
	FavoriteArtists []string `json:"favorite_artists"`
	FavoriteAlbums  []string `json:"favorite_albums"`
	IsConnected     bool
}

type Tracks struct {
	Total int     `json:"total"`
	Items []Track `json:"data"`
}

type Track struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Release string `json:"release_date"`
	Album   Album  `json:"album"`
	Artist  Artist `json:"artist"`
}

type Album struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	Tracklist string `json:"tracklist"`
}

type Albums struct {
	Total int     `json:"total"`
	Items []Album `json:"data"`
}

type Artist struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Nb_albums int    `json:"nb_album"`
	TrackList string `json:"tracklist"`
}

type Artists struct {
	Total int      `json:"total"`
	Items []Artist `json:"data"`
}

func main() {
	temp, err := template.ParseGlob("assets/templates/*.html")
	if err != nil {
		fmt.Println("Erreur dans la récupération des templates : ", err)
		return
	}

	CurrentUser := User{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "index", CurrentUser)
	})

	http.HandleFunc("/tracks", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/track/?q=Dethklok"

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Problème danss la requête d'obtention des tracks : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans l'envoi de la requête d'obtention des tracks : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans la lecture de la réponse de tracks : ", errBody)
		}

		var decodeData Tracks
		json.Unmarshal(body, &decodeData)

		var Data []Track
		for i := 0; i < len(decodeData.Items); i++ {
			Data = append(Data, decodeData.Items[i])
		}

		temp.ExecuteTemplate(w, "tracks", Data)
	})

	http.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/artist"

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Problème danss la requête d'obtention des artistes : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans l'envoi de la requête d'obtention des artistes : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans la lecture de la réponse de artistes : ", errBody)
		}

		var decodeData Artists
		json.Unmarshal(body, &decodeData)

		var Data []Artist
		for i := 0; i < len(decodeData.Items); i++ {
			Data = append(Data, decodeData.Items[i])
		}

		temp.ExecuteTemplate(w, "artists", Data)
	})

	http.HandleFunc("/ablums", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/album"

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Problème danss la requête d'obtention des albums : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans l'envoi de la requête d'obtention des albums : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans la lecture de la réponse de Albums : ", errBody)
		}

		var decodeData Albums
		json.Unmarshal(body, &decodeData)

		var Data []Album
		for i := 0; i < len(decodeData.Items); i++ {
			Data = append(Data, decodeData.Items[i])
		}

		temp.ExecuteTemplate(w, "albums", Data)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {

		temp.ExecuteTemplate(w, "search", CurrentUser)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		filter := r.PostFormValue("filter")
		research := r.PostFormValue("search")
		api_url := "http://api.deezer.com/search" + filter + "?q=" + research

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Problème danss la requête de recherche : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans l'envoi de la requête de recherche : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans la lecture de la réponse de recherche : ", errBody)
		}

		decodeTrack := Tracks{}
		decodeArtist := Artists{}
		decodeAlbum := Albums{}

		var DataTrack []Track
		var DataAlbum []Album
		var DataArtist []Artist

		if filter == "artist" {
			json.Unmarshal(body, &decodeArtist)

			for i := 0; i < len(decodeArtist.Items); i++ {
				DataArtist = append(DataArtist, decodeArtist.Items[i])
			}
			temp.ExecuteTemplate(w, "result", DataArtist)
		} else if filter == "album" {
			json.Unmarshal(body, &decodeAlbum)

			for i := 0; i < len(decodeAlbum.Items); i++ {
				DataAlbum = append(DataAlbum, decodeAlbum.Items[i])
			}

			temp.ExecuteTemplate(w, "result", DataAlbum)
		} else {
			json.Unmarshal(body, &decodeTrack)

			for i := 0; i < len(decodeTrack.Items); i++ {
				DataTrack = append(DataTrack, decodeTrack.Items[i])
			}

			temp.ExecuteTemplate(w, "result", DataTrack)
		}
	})

	http.HandleFunc("/a_propos", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "a_propos", CurrentUser)
	})

	http.HandleFunc("/favorites", func(w http.ResponseWriter, r *http.Request) {

	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {

	})
	http.HandleFunc("/connectHandler", func(w http.ResponseWriter, r *http.Request) {

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

	})

	RootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(RootDoc + "/assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	http.ListenAndServe("localhost:8080", nil)
}
