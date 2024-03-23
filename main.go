package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type User struct {
	UserName        string   `json:"username"`
	Password        string   `json:"password"`
	FavoriteTracks  []Track  `json:"favorite_tracks"`
	FavoriteArtists []Artist `json:"favorite_artists"`
	FavoriteAlbums  []Album  `json:"favorite_albums"`
}

type Tracks struct {
	Total int     `json:"total"`
	Items []Track `json:"data"`
}

type Track struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Release string `json:"release_date"`
	Album   Album  `json:"album"`
	Artist  Artist `json:"artist"`
}

type Album struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	Tracklist string `json:"tracklist"`
	Artist    Artist `json:"artist"`
}

type Albums struct {
	Total int     `json:"total"`
	Items []Album `json:"data"`
}

type Artist struct {
	Id        int    `json:"id"`
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
	IsConnected := false
	temp, err := template.ParseGlob("assets/templates/*.html")
	if err != nil {
		fmt.Println("Erreur dans la récupération des templates : ", err)
		return
	}

	CurrentUser := User{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected bool
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected
		temp.ExecuteTemplate(w, "index", Data)
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

		var Data struct {
			Results   []Track
			User      User
			Connected bool
		}
		for i := 0; i < len(decodeData.Items); i++ {
			Data.Results = append(Data.Results, decodeData.Items[i])
		}
		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "tracks", Data)
	})

	http.HandleFunc("/track", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RawQuery[2:]

		id, _ := strconv.Atoi(url)

		api_url := "http://api.deezer.com/track/" + url

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}

		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la requête pour le track : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour le track : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse du track : ", errBody)
		}

		var DecodeData Track
		json.Unmarshal(body, &DecodeData)

		var Data struct {
			Result     Track
			User       User
			AlreadyFav bool
			Connected  bool
		}

		Data.Result = DecodeData
		Data.User = CurrentUser
		Data.AlreadyFav = false
		Data.Connected = IsConnected
		if IsConnected {
			for i := 0; i < len(CurrentUser.FavoriteTracks); i++ {
				if id == CurrentUser.FavoriteTracks[i].Id {
					Data.AlreadyFav = true
				}
			}
		}

		temp.ExecuteTemplate(w, "song", Data)
	})
	http.HandleFunc("/AddToFav/Track", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RawQuery[2:]

		id, _ := strconv.Atoi(url)

		if IsConnected {
			for i := 0; i < len(CurrentUser.FavoriteTracks); i++ {
				if id == CurrentUser.FavoriteTracks[i].Id {
					http.Redirect(w, r, "/track?t="+url, http.StatusSeeOther)
				}
			}

			api_url := "http://api.deezer.com/track/" + url

			httpClient := http.Client{
				Timeout: time.Second * 2,
			}

			req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
			if errReq != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la requête pour le track : ", errReq)
			}

			res, errRes := httpClient.Do(req)
			if res.Body != nil {
				defer res.Body.Close()
			} else {
				fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour le track : ", errRes)
			}

			body, errBody := io.ReadAll(res.Body)
			if errBody != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse du track : ", errBody)
			}

			var DecodeData Track
			json.Unmarshal(body, &DecodeData)

			CurrentUser.FavoriteTracks = append(CurrentUser.FavoriteTracks, DecodeData)
			fmt.Println(CurrentUser.FavoriteTracks)
			Json, err := os.ReadFile("./user.json")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la lecture du fichier user.json : ", err)
			}

			var allUsers struct {
				Users []User
			}
			json.Unmarshal(Json, &allUsers)

			for i := 0; i < len(allUsers.Users); i++ {
				if allUsers.Users[i].UserName == CurrentUser.UserName {
					allUsers.Users[i] = CurrentUser
				}
			}

			newJSON, err := json.MarshalIndent(allUsers, "", "    ")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la création des nouvelles données json : ", err)
			}

			err = os.WriteFile("./user.json", newJSON, 0644)
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans l'écriture du nouveau json : ", err)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/connect", http.StatusSeeOther)
		}
	})
	http.HandleFunc("/AddToFav/Album", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RawQuery[2:]

		id, _ := strconv.Atoi(url)

		if IsConnected {
			for i := 0; i < len(CurrentUser.FavoriteAlbums); i++ {
				if id == CurrentUser.FavoriteAlbums[i].Id {
					http.Redirect(w, r, "/album?a="+url, http.StatusSeeOther)
				}
			}

			api_url := "http://api.deezer.com/album/" + url

			httpClient := http.Client{
				Timeout: time.Second * 2,
			}

			req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
			if errReq != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la requête pour le track : ", errReq)
			}

			res, errRes := httpClient.Do(req)
			if res.Body != nil {
				defer res.Body.Close()
			} else {
				fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour le track : ", errRes)
			}

			body, errBody := io.ReadAll(res.Body)
			if errBody != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse du track : ", errBody)
			}

			var DecodeData Album
			json.Unmarshal(body, &DecodeData)

			CurrentUser.FavoriteAlbums = append(CurrentUser.FavoriteAlbums, DecodeData)
			Json, err := os.ReadFile("./user.json")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la lecture du fichier user.json : ", err)
			}

			var allUsers struct {
				Users []User
			}
			json.Unmarshal(Json, &allUsers)

			for i := 0; i < len(allUsers.Users); i++ {
				if allUsers.Users[i].UserName == CurrentUser.UserName {
					allUsers.Users[i] = CurrentUser
				}
			}

			newJSON, err := json.MarshalIndent(allUsers, "", "    ")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la création des nouvelles données json : ", err)
			}

			err = os.WriteFile("./user.json", newJSON, 0644)
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans l'écriture du nouveau json : ", err)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/connect", http.StatusSeeOther)
		}
	})
	http.HandleFunc("/AddToFav/Artist", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RawQuery[2:]

		id, _ := strconv.Atoi(url)

		if IsConnected {
			for i := 0; i < len(CurrentUser.FavoriteArtists); i++ {
				if id == CurrentUser.FavoriteArtists[i].Id {
					http.Redirect(w, r, "/artist?a="+url, http.StatusSeeOther)
				}
			}

			api_url := "http://api.deezer.com/artist/" + url

			httpClient := http.Client{
				Timeout: time.Second * 2,
			}

			req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
			if errReq != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la requête pour le track : ", errReq)
			}

			res, errRes := httpClient.Do(req)
			if res.Body != nil {
				defer res.Body.Close()
			} else {
				fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour le track : ", errRes)
			}

			body, errBody := io.ReadAll(res.Body)
			if errBody != nil {
				fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse du track : ", errBody)
			}

			var DecodeData Artist
			json.Unmarshal(body, &DecodeData)

			CurrentUser.FavoriteArtists = append(CurrentUser.FavoriteArtists, DecodeData)
			fmt.Println(CurrentUser.FavoriteArtists)
			Json, err := os.ReadFile("./user.json")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la lecture du fichier user.json : ", err)
			}

			var allUsers struct {
				Users []User
			}
			json.Unmarshal(Json, &allUsers)

			for i := 0; i < len(allUsers.Users); i++ {
				if allUsers.Users[i].UserName == CurrentUser.UserName {
					allUsers.Users[i] = CurrentUser
				}
			}

			newJSON, err := json.MarshalIndent(allUsers, "", "    ")
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans la création des nouvelles données json : ", err)
			}

			err = os.WriteFile("./user.json", newJSON, 0644)
			if err != nil {
				fmt.Println("Erreur dans le toaster ! Problème dans l'écriture du nouveau json : ", err)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/connect", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/album", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.RawQuery[2:]

		api_url := "http://api.deezer.com/album/" + id

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}

		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la requête pour l'album : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour l'album : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse de l'album : ", errBody)
		}

		var DecodeData Album
		json.Unmarshal(body, &DecodeData)

		var Data struct {
			Result    Album
			User      User
			Connected bool
		}

		Data.Result = DecodeData
		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "album", Data)
	})

	http.HandleFunc("/artist", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.RawQuery[2:]

		api_url := "http://api.deezer.com/artist/" + id

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}

		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la requête pour l'artiste : ", errReq)
		}

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour l'artiste : ", errRes)
		}

		body, errBody := io.ReadAll(res.Body)
		if errBody != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse de l'artiste : ", errBody)
		}

		var DecodeData Artist
		json.Unmarshal(body, &DecodeData)

		var Data struct {
			Result    Artist
			User      User
			Connected bool
		}

		Data.Result = DecodeData
		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "artist", Data)
	})
	http.HandleFunc("/artists", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/artist?q=Dethklok"

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

		var Data struct {
			Results   []Artist
			User      User
			Connected bool
		}
		for i := 0; i < len(decodeData.Items); i++ {
			Data.Results = append(Data.Results, decodeData.Items[i])
		}
		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "artists", Data)
	})

	http.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/album?q=Dethklok"

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

		var Data struct {
			Results   []Album
			User      User
			Connected bool
		}
		for i := 0; i < len(decodeData.Items); i++ {
			Data.Results = append(Data.Results, decodeData.Items[i])
		}
		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "albums", Data)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected bool
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "search", Data)
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
			fmt.Println("Problème dans la requête de recherche : ", errReq)
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

		var DataTrack struct {
			Results   []Track
			User      User
			Filter    string
			Connected bool
		}
		var DataAlbum struct {
			Results   []Album
			User      User
			Filter    string
			Connected bool
		}
		var DataArtist struct {
			Results   []Artist
			User      User
			Filter    string
			Connected bool
		}

		if filter == "artists" {
			json.Unmarshal(body, &decodeArtist)

			for i := 0; i < len(decodeArtist.Items); i++ {
				DataArtist.Results = append(DataArtist.Results, decodeArtist.Items[i])
			}
			DataArtist.User = CurrentUser
			DataArtist.Filter = "artist"
			DataArtist.Connected = IsConnected

			temp.ExecuteTemplate(w, "result", DataArtist)
		} else if filter == "album" {
			json.Unmarshal(body, &decodeAlbum)

			for i := 0; i < len(decodeAlbum.Items); i++ {
				DataAlbum.Results = append(DataAlbum.Results, decodeAlbum.Items[i])
			}
			DataAlbum.User = CurrentUser
			DataAlbum.Filter = "album"
			DataAlbum.Connected = IsConnected

			temp.ExecuteTemplate(w, "result", DataAlbum)
		} else {
			json.Unmarshal(body, &decodeTrack)

			for i := 0; i < len(decodeTrack.Items); i++ {
				DataTrack.Results = append(DataTrack.Results, decodeTrack.Items[i])
			}
			DataTrack.User = CurrentUser
			DataTrack.Filter = "track"
			DataTrack.Connected = IsConnected

			temp.ExecuteTemplate(w, "result", DataTrack)
		}
	})

	http.HandleFunc("/a_propos", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected bool
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "a_propos", Data)
	})

	http.HandleFunc("/favorites", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected bool
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		fmt.Println(IsConnected)

		if IsConnected {
			temp.ExecuteTemplate(w, "favoris", Data)
		}
		temp.ExecuteTemplate(w, "connect", Data)
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		if IsConnected {
			http.Redirect(w, r, "/", http.StatusForbidden)
		} else {
			type Error struct {
				Err       string
				User      User
				Connected bool
			}
			url := r.URL.RawQuery
			switch url {
			case "err=not_exists":
				err := Error{
					Err:       "Not Exists",
					User:      CurrentUser,
					Connected: IsConnected,
				}
				temp.ExecuteTemplate(w, "connection", err)
			case "err=w_pwd":
				err := Error{
					Err:       "Wrong Password",
					User:      CurrentUser,
					Connected: IsConnected,
				}
				temp.ExecuteTemplate(w, "connection", err)
			default:
				err := Error{
					Err:       "nul",
					User:      CurrentUser,
					Connected: IsConnected,
				}
				temp.ExecuteTemplate(w, "connection", err)
			}
		}

	})
	http.HandleFunc("/connectHandler", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("username")
		pwd := r.FormValue("password")

		userData, userErr := os.ReadFile("user.json")
		if userErr != nil {
			fmt.Println("Erreur dans la lecture du fichier user : ", userErr)
		}

		var allUsers struct {
			Users []User
		}
		json.Unmarshal(userData, &allUsers)
		wrongPwd := false
		exists := false
		type Favorites struct {
			Albums  []Album
			Tracks  []Track
			Artists []Artist
		}
		var userFav Favorites

		for i := 0; i < len(allUsers.Users); i++ {
			if name == allUsers.Users[i].UserName {
				exists = true
				if pwd != allUsers.Users[i].Password {
					wrongPwd = true
				} else {
					userFav.Albums = allUsers.Users[i].FavoriteAlbums
					userFav.Tracks = allUsers.Users[i].FavoriteTracks
					userFav.Artists = allUsers.Users[i].FavoriteArtists
				}
			}
		}

		if wrongPwd && exists {
			http.Redirect(w, r, "/connect?err=w_pwd", http.StatusSeeOther)
		} else if !exists {
			http.Redirect(w, r, "/connect?err=not_exists", http.StatusSeeOther)
		} else {
			IsConnected = true
			CurrentUser = User{
				UserName:        name,
				Password:        pwd,
				FavoriteTracks:  userFav.Tracks,
				FavoriteArtists: userFav.Artists,
				FavoriteAlbums:  userFav.Albums,
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if IsConnected {
			http.Redirect(w, r, "/", http.StatusForbidden)
		} else {
			url := r.URL.RawQuery
			var Data struct {
				Err       string
				User      User
				Connected bool
			}

			if url == "err=alreadyUsed" {
				Data.Err = "alreadyUsed"
			} else {
				Data.Err = "nul"
			}
			Data.User = CurrentUser
			Data.Connected = IsConnected

			temp.ExecuteTemplate(w, "register", Data)
		}
	})
	http.HandleFunc("/registerHandler", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("username")
		pwd := r.FormValue("password")

		userData, userErr := os.ReadFile("./user.json")
		if userErr != nil {
			fmt.Println("Erreur dans la lecture du fichier user : ", userErr)
		}

		var allUsers struct {
			Users []User
		}
		json.Unmarshal(userData, &allUsers)

		alreadyUsed := false

		for i := 0; i < len(allUsers.Users); i++ {
			if name == allUsers.Users[i].UserName { //si le nom a déjà été utilisé
				alreadyUsed = true
			}
		}

		if alreadyUsed {
			http.Redirect(w, r, "/register?err=alreadyUsed", http.StatusBadRequest)
		} else {
			allUsers.Users = append(allUsers.Users, User{
				UserName:        name,
				Password:        pwd,
				FavoriteTracks:  []Track{},
				FavoriteAlbums:  []Album{},
				FavoriteArtists: []Artist{},
			})

			newJSON, err := json.MarshalIndent(allUsers, "", "   ")

			if err != nil {
				fmt.Println("Erreur dans la conversion du nouveau compte en json : ", err)
			}

			err = os.WriteFile("./user.json", newJSON, 0644)
			if err != nil {
				fmt.Println("Erreur dans l'écriture du json : ", err)
			}

			http.Redirect(w, r, "/connect", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/deconnect", func(w http.ResponseWriter, r *http.Request) {
		CurrentUser = User{
			UserName:        "",
			Password:        "",
			FavoriteTracks:  []Track{},
			FavoriteAlbums:  []Album{},
			FavoriteArtists: []Artist{},
		}

		IsConnected = false

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	RootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(RootDoc + "/assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	http.ListenAndServe("localhost:8080", nil)
}

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
