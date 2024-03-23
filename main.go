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
	TrackList string `json:"tracklist"`
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
	IsConnected := "false"
	temp, err := template.ParseGlob("assets/templates/*.html")
	if err != nil {
		fmt.Println("Erreur dans la récupération des templates : ", err)
		return
	}

	CurrentUser := User{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected string
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected
		temp.ExecuteTemplate(w, "index", Data)
	})

	http.HandleFunc("/tracks", func(w http.ResponseWriter, r *http.Request) {
		api_url := "http://api.deezer.com/search/track/?q=Dethklok&limit=10"

		httpClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, errReq := http.NewRequest(http.MethodGet, api_url, nil)
		if errReq != nil {
			fmt.Println("Problème dans la requête d'obtention des tracks : ", errReq)
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
			Connected string
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
			Connected  string
		}

		Data.Result = DecodeData
		Data.User = CurrentUser
		Data.AlreadyFav = false
		Data.Connected = IsConnected
		if IsConnected == "true" {
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

		if IsConnected == "true" {
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

		if IsConnected == "true" {
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

		if IsConnected == "true" {
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

		var Tracklist struct {
			Tracks []Track `json:"data"`
		}

		trackReq, errReq := http.NewRequest(http.MethodGet, DecodeData.TrackList, nil)
		if errReq != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la requête pour la trackilst : ", errReq)
		}

		trackRes, errRes := httpClient.Do(trackReq)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour la trackilst : ", errRes)
		}

		trackBody, errBody := io.ReadAll(trackRes.Body)
		if errBody != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse de la tracklist : ", errBody)
		}

		json.Unmarshal(trackBody, &Tracklist)

		var Data struct {
			Result    Album
			Tracklist struct {
				Tracks []Track `json:"data"`
			}
			User      User
			Connected string
		}

		Data.Result = DecodeData
		Data.Tracklist = Tracklist
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

		var Tracklist struct {
			Tracks []Track
		}

		trackReq, errReq := http.NewRequest(http.MethodGet, DecodeData.TrackList, nil)
		if errReq != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la requête pour la trackilst : ", errReq)
		}

		trackRes, errRes := httpClient.Do(trackReq)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Erreur dans le toaster! Problème dans l'envoi de la requête pour la trackilst : ", errRes)
		}

		trackBody, errBody := io.ReadAll(trackRes.Body)
		if errBody != nil {
			fmt.Println("Erreur dans le toaster! Problème dans la lecture du body de la réponse de la tracklist : ", errBody)
		}

		json.Unmarshal(trackBody, &Tracklist)

		var Data struct {
			Result    Artist
			TrackList struct {
				Tracks []Track
			}
			User      User
			Connected string
		}

		Data.Result = DecodeData
		Data.User = CurrentUser
		Data.Connected = IsConnected
		Data.TrackList = Tracklist

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
			Connected string
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
			Connected string
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
			Connected string
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "search", Data)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		research := r.URL.Query().Get("search")
		filter := r.URL.Query().Get("filter")
		api_url := "http://api.deezer.com/search/" + filter + "?q=" + research

		page := r.URL.Query().Get("page")

		ipage, _ := strconv.Atoi(page)

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

		var DataTrack struct {
			Results []Track `json:"data"`
		}
		var DataAlbum struct {
			Results []Album `json:"data"`
		}
		var DataArtist struct {
			Results []Artist `json:"data"`
		}

		if filter == "artists" {
			json.Unmarshal(body, &DataArtist)

			var Data struct {
				Data struct {
					Results []Artist
				}
				User      User
				Connected string
				Page      int
				PageSuiv  int
				PagePrec  int
				LastPage  int
				Filter    string
				Search    string
			}
			Data.Connected = IsConnected
			Data.Page = ipage
			Data.Search = research
			Data.Filter = filter
			Data.LastPage = len(DataArtist.Results) / 10
			if len(DataArtist.Results) > ipage*10 {
				Data.PageSuiv = Data.Page + 1
			} else {
				Data.PageSuiv = Data.Page
			}
			if ipage == 0 {
				if len(DataArtist.Results) >= 9 {
					Data.Data.Results = DataArtist.Results[:10]
				} else {
					Data.Data.Results = DataArtist.Results
				}
				Data.PagePrec = 0
			} else {
				if len(DataArtist.Results) >= ipage*10+10 {
					Data.Data.Results = DataArtist.Results[ipage*10 : ipage*10+10]
				} else {
					Data.Data.Results = DataArtist.Results[ipage*10:]
				}
				Data.PagePrec = Data.Page - 1
			}

			temp.ExecuteTemplate(w, "result", Data)
		} else if filter == "album" {
			json.Unmarshal(body, &DataAlbum)

			var Data struct {
				Data struct {
					Results []Album
				}
				User      User
				Connected string
				Page      int
				PageSuiv  int
				PagePrec  int
				LastPage  int
				Filter    string
				Search    string
			}
			Data.Connected = IsConnected
			Data.Page = ipage
			Data.Search = research
			Data.Filter = filter
			Data.LastPage = len(DataAlbum.Results) / 10
			if len(DataAlbum.Results) > ipage*10 {
				Data.PageSuiv = Data.Page + 1
			}
			if ipage == 0 {
				if len(DataAlbum.Results) >= 9 {
					Data.Data.Results = DataAlbum.Results[:10]
				} else {
					Data.Data.Results = DataAlbum.Results
				}
				Data.PagePrec = 0
			} else {
				if len(DataArtist.Results) >= ipage*10+10 {
					Data.Data.Results = DataAlbum.Results[ipage*10 : ipage*10+10]
					Data.PagePrec = Data.Page - 1
				} else {
					Data.Data.Results = DataAlbum.Results[ipage*10:]
				}
			}

			temp.ExecuteTemplate(w, "result", Data)
		} else {
			json.Unmarshal(body, &DataTrack)

			fmt.Println("passed")
			var Data struct {
				Data struct {
					Results []Track
				}
				User      User
				Connected string
				Page      int
				PageSuiv  int
				PagePrec  int
				LastPage  int
				Filter    string
				Search    string
			}
			Data.Connected = IsConnected
			Data.Page = ipage
			Data.Filter = filter
			Data.Search = research
			Data.LastPage = len(DataTrack.Results) / 10
			if len(DataTrack.Results) > ipage*10 {
				Data.PageSuiv = Data.Page + 1
			}
			if ipage == 0 {
				if len(DataTrack.Results) >= 9 {
					Data.Data.Results = DataTrack.Results[:10]
				} else {
					Data.Data.Results = DataTrack.Results
				}
				Data.PagePrec = 0
			} else {
				if len(DataTrack.Results) >= ipage*10+10 {
					Data.Data.Results = DataTrack.Results[ipage*10 : ipage*10+10]
				} else {
					Data.Data.Results = DataTrack.Results[ipage*10:]
				}
				Data.PagePrec = Data.Page - 1
			}

			temp.ExecuteTemplate(w, "result", Data)
		}
	})

	http.HandleFunc("/a_propos", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected string
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		temp.ExecuteTemplate(w, "a_propos", Data)
	})

	http.HandleFunc("/favorites", func(w http.ResponseWriter, r *http.Request) {
		var Data struct {
			User      User
			Connected string
		}

		Data.User = CurrentUser
		Data.Connected = IsConnected

		fmt.Println(IsConnected)

		if IsConnected == "true" {
			temp.ExecuteTemplate(w, "favoris", Data)
		}
		http.Redirect(w, r, "/connect", http.StatusSeeOther)
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		if IsConnected == "true" {
			http.Redirect(w, r, "/", http.StatusForbidden)
		} else {
			type Error struct {
				Err       string
				User      User
				Connected string
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
			IsConnected = "true"
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
		if IsConnected == "true" {
			http.Redirect(w, r, "/", http.StatusForbidden)
		} else {
			url := r.URL.RawQuery
			var Data struct {
				Err       string
				User      User
				Connected string
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

		IsConnected = "false"

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	RootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(RootDoc + "/assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	http.ListenAndServe("localhost:8080", nil)
}
