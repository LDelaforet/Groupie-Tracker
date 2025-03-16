package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Normalement ca ca bouge pas mais je rends dynamique au cas ou
var apiURL = ""
var user_agent = ""
var contentType = ""
var callbackURL = ""

var oauthStringEmpty = "OAuth " +
	"oauth_consumer_key=\"%s\", oauth_nonce=\"%s\"," +
	"oauth_token=\"%s\", oauth_signature=\"%s&%s\", " +
	"oauth_signature_method=\"PLAINTEXT\", oauth_timestamp=\"%d\", " +
	"oauth_verifier=\"%s\""

// A remplacer par celles de l'utilisateur
var consumerKey = ""
var consumerSecret = ""

// Possiblement sensible a la session, a vérifier
var oauth_token = ""
var oauth_token_secret = ""
var oauth_verifier = ""

var oauthLink = ""

var host = ""
var port = 0

var authentified = false

var favorites = make(map[string]Result)
var favoritesFile = "../data/favoris.json"

var temp *template.Template

// Structure pour les détails d'un artiste
var pageVars struct {
	OAuthLink  string
	OAuthPopup bool
	Artist     ArtistDetail
	Name       string
	RealName   string
	URI        string
}

type Favorite struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	ImageLink  string `json:"cover_image"`
	Country    string `json:"country"`
	Year       string `json:"year"`
	TypeDisp   string `json:"TypeDisp"`
	IsFavorite bool   `json:"IsFavorite"`
}

type Result struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	ImageLink  string `json:"cover_image"`
	Country    string `json:"country"`
	Year       string `json:"year"`
	TypeDisp   string
	IsFavorite bool `json:"IsFavorite"`
}

type SearchResultPvars struct {
	OAuthLink   string
	OAuthPopup  bool
	Results     []Result `json:"results"`
	CurrentPage int
	PerPage     int
	HasNextPage bool
	SearchQuery string
	Year        string
	Type        string
	Genre       string
}

type FavoritesPagePvars struct {
	OAuthLink  string
	OAuthPopup bool
	Favorites  []Result `json:"favorites"`
}

type PassedVars struct {
	OAuthLink  string
	OAuthPopup bool
}

type ArtistReleasesPageVars struct {
	OAuthLink   string
	OAuthPopup  bool
	ArtistName  string
	ArtistID    string
	Releases    []Release
	CurrentPage int
	PerPage     int
	HasNextPage bool
	Type        string
	Year        string
}

type Release struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Year       int    `json:"year"`
	Thumb      string `json:"thumb"`
	IsFavorite bool   `json:"is_favorite"`
	TypeDisp   string
}

type ArtistReleasesResponse struct {
	OAuthLink  string
	OAuthPopup bool
	Pagination struct {
		Page    int `json:"page"`
		Pages   int `json:"pages"`
		PerPage int `json:"per_page"`
		Items   int `json:"items"`
		Urls    struct {
			Last string `json:"last"`
			Next string `json:"next"`
		} `json:"urls"`
	} `json:"pagination"`
	Releases   []Release `json:"releases"`
	ArtistName string
	ArtistID   string
	Type       string
	Year       string
}

// Genere un nonce psq apparament on est en OAuth 1.0 mais bon ca marche tt autant...
func generateNonce() string {
	nonce := make([]byte, 16)
	rand.Read(nonce)
	return hex.EncodeToString(nonce)
}

// Pour faire la string Authorization du header
func OauthCrafter() string {
	timestamp := time.Now().Unix()
	nonce := generateNonce()
	oauthString := fmt.Sprintf(
		oauthStringEmpty,
		consumerKey, nonce, oauth_token, consumerSecret, oauth_token_secret, timestamp, oauth_verifier)
	return oauthString
}

func addproduct(w http.ResponseWriter, r *http.Request) {
}

func isAuthorized() bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/oauth/identity", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	defer resp.Body.Close()

	authentified = resp.StatusCode == http.StatusOK

	return resp.StatusCode == http.StatusOK
}

func requestOauthToken() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/oauth/request_token?oauth_callback="+callbackURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: %v", resp.Status)
	}

	// La string ressemble a ca: oauth_token=xxxx&oauth_token_secret=xxxx
	// On veut recup seulement les xxxx
	oauth_token = strings.Split(strings.Split(string(body), "&")[0], "=")[1]
	oauth_token_secret = strings.Split(strings.Split(string(body), "&")[1], "=")[1]

	return nil
}

func accessOauthToken() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/oauth/access_token", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	oauth_token = strings.Split(strings.Split(string(body), "&")[0], "=")[1]
	oauth_token_secret = strings.Split(strings.Split(string(body), "&")[1], "=")[1]

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: %v", resp.Status)
	}

	return nil
}

func getSettings() {
	configFile, err := os.ReadFile("../config/config.json")
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier de configuration:", err)
		return
	}

	var config map[string]interface{}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("Erreur lors du parsing du fichier de configuration:", err)
		return
	}
	if apiConfig, ok := config["api"].(map[string]interface{}); ok {
		apiURL = apiConfig["url"].(string)
		user_agent = apiConfig["user_agent"].(string)
		contentType = apiConfig["content_type"].(string)
	}

	if oauthConfig, ok := config["oauth"].(map[string]interface{}); ok {
		consumerKey = oauthConfig["consumer_key"].(string)
		consumerSecret = oauthConfig["consumer_secret"].(string)
		callbackURL = oauthConfig["callback_url"].(string)
	}

	if serverConfig, ok := config["server"].(map[string]interface{}); ok {
		host = serverConfig["host"].(string)
		port = int(serverConfig["port"].(float64))
	}
}

func searchAPI(query string, year string, genre string, Type string, page int, perPage int) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/database/search?q="+query+"&type="+Type+"&genre="+genre+"&year="+year+"&per_page="+strconv.Itoa(perPage)+"&page="+strconv.Itoa(page), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %v", resp.Status)
	}

	return body, nil
}

func getRelease(releaseID string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/releases/"+releaseID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %v", resp.Status)
	}

	return body, nil
}

// Fonction pour sauvegarder les favoris dans le fichier JSON
func saveFavorites() error {
	// Créer le dossier data s'il n'existe pas
	os.MkdirAll("../data", 0755)

	data, err := json.MarshalIndent(favorites, "", "    ")
	if err != nil {
		return fmt.Errorf("erreur lors de la conversion en JSON: %v", err)
	}

	err = ioutil.WriteFile(favoritesFile, data, 0644)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture du fichier: %v", err)
	}

	return nil
}

// Fonction pour charger les favoris depuis le fichier JSON
func loadFavorites() error {
	data, err := ioutil.ReadFile(favoritesFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Créer le dossier data s'il n'existe pas
			os.MkdirAll("../data", 0755)

			// Créer un fichier JSON vide avec {}
			emptyJSON := []byte("{}")
			if err := ioutil.WriteFile(favoritesFile, emptyJSON, 0644); err != nil {
				return fmt.Errorf("erreur lors de la création du fichier JSON vide: %v", err)
			}

			favorites = make(map[string]Result)
			return nil
		}
		return fmt.Errorf("erreur lors de la lecture du fichier: %v", err)
	}

	// Si le fichier est vide, initialiser avec {}
	if len(data) == 0 {
		emptyJSON := []byte("{}")
		if err := ioutil.WriteFile(favoritesFile, emptyJSON, 0644); err != nil {
			return fmt.Errorf("erreur lors de l'écriture du fichier JSON vide: %v", err)
		}
		favorites = make(map[string]Result)
		return nil
	}

	err = json.Unmarshal(data, &favorites)
	if err != nil {
		// Si erreur de parsing, tenter de réinitialiser le fichier
		emptyJSON := []byte("{}")
		if err := ioutil.WriteFile(favoritesFile, emptyJSON, 0644); err != nil {
			return fmt.Errorf("erreur lors de la réinitialisation du fichier JSON: %v", err)
		}
		favorites = make(map[string]Result)
		return nil
	}

	return nil
}

// Structures pour les détails d'une release
type Artist struct {
	Name         string `json:"name"`
	ID           int    `json:"id"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type Label struct {
	Name         string `json:"name"`
	ID           int    `json:"id"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type Format struct {
	Name         string   `json:"name"`
	Qty          string   `json:"qty"`
	Descriptions []string `json:"descriptions"`
}

type Track struct {
	Position string `json:"position"`
	Type     string `json:"type_"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
}

type Image struct {
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type ReleaseDetail struct {
	ID         int      `json:"id"`
	Year       int      `json:"year"`
	URI        string   `json:"uri"`
	Artists    []Artist `json:"artists"`
	Labels     []Label  `json:"labels"`
	Formats    []Format `json:"formats"`
	MasterID   int      `json:"master_id"`
	Title      string   `json:"title"`
	Country    string   `json:"country"`
	Released   string   `json:"released"`
	Genres     []string `json:"genres"`
	Styles     []string `json:"styles"`
	Tracklist  []Track  `json:"tracklist"`
	Images     []Image  `json:"images"`
	OAuthLink  string
	OAuthPopup bool
}

type ArtistDetail struct {
	Name       string  `json:"name"`
	ID         int     `json:"id"`
	URI        string  `json:"uri"`
	Images     []Image `json:"images"`
	RealName   string  `json:"realname"`
	Profile    string  `json:"profile"`
	OAuthLink  string
	OAuthPopup bool
}

// Structure pour représenter les détails d'un master
type MasterDetail struct {
	ID                int      `json:"id"`
	MainRelease       int      `json:"main_release"`
	MostRecentRelease int      `json:"most_recent_release"`
	URI               string   `json:"uri"`
	Genres            []string `json:"genres"`
	Styles            []string `json:"styles"`
	Year              int      `json:"year"`
	Tracklist         []Track  `json:"tracklist"`
	Artists           []Artist `json:"artists"`
	Title             string   `json:"title"`
	Notes             string   `json:"notes"`
	Images            []Image  `json:"images"` // Sera rempli à partir de la main_release
	Country           string   `json:"country"`
	Released          string   `json:"released"`
	Formats           []Format `json:"formats"`
	Labels            []Label  `json:"labels"`
	OAuthLink         string
	OAuthPopup        bool
}

// Fonction pour récupérer les détails d'une release
func getReleaseDetails(releaseID string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/releases/"+releaseID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %v", resp.Status)
	}

	return body, nil
}

// Fonction pour récupérer les détails d'un master
func getMasterDetails(masterID string) (*MasterDetail, error) {
	// Récupérer les détails du master
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL+"/masters/"+masterID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %v", resp.Status)
	}

	// Désérialiser les données JSON du master
	var masterDetail MasterDetail
	if err := json.Unmarshal(body, &masterDetail); err != nil {
		return nil, err
	}

	// Récupérer les détails de la release principale pour obtenir les images
	mainReleaseID := strconv.Itoa(masterDetail.MainRelease)

	// Récupérer les détails de la release principale
	client = &http.Client{}
	req, err = http.NewRequest("GET", apiURL+"/releases/"+mainReleaseID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	releaseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %v", resp.Status)
	}

	// Désérialiser les données JSON de la release principale
	var releaseDetail ReleaseDetail
	if err := json.Unmarshal(releaseBody, &releaseDetail); err != nil {
		return nil, err
	}

	// Ajouter les images de la release principale au master
	masterDetail.Images = releaseDetail.Images

	// Ajouter les autres informations de la release principale au master
	masterDetail.Country = releaseDetail.Country
	masterDetail.Released = releaseDetail.Released
	masterDetail.Formats = releaseDetail.Formats
	masterDetail.Labels = releaseDetail.Labels

	return &masterDetail, nil
}

func getArtistReleases(artistID string, page int, perPage int, releaseType string, year string) (ArtistReleasesResponse, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/artists/%s/releases?page=%d&per_page=%d", apiURL, artistID, page, perPage)
	if releaseType != "" {
		url += "&type=" + releaseType
	}
	if year != "" {
		url += "&year=" + year
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ArtistReleasesResponse{}, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return ArtistReleasesResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArtistReleasesResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return ArtistReleasesResponse{}, fmt.Errorf("Error: %v", resp.Status)
	}

	var artistReleasesResponse ArtistReleasesResponse
	if err := json.Unmarshal(body, &artistReleasesResponse); err != nil {
		return ArtistReleasesResponse{}, err
	}

	// Vérifier si chaque release est dans les favoris et définir TypeDisp
	for i := range artistReleasesResponse.Releases {
		releaseID := strconv.Itoa(artistReleasesResponse.Releases[i].ID)
		if _, exists := favorites[releaseID]; exists {
			artistReleasesResponse.Releases[i].IsFavorite = true
		}
		// Définir TypeDisp
		switch artistReleasesResponse.Releases[i].Type {
		case "release":
			artistReleasesResponse.Releases[i].TypeDisp = "Sortie"
		case "master":
			artistReleasesResponse.Releases[i].TypeDisp = "Master"
		case "artist":
			artistReleasesResponse.Releases[i].TypeDisp = "Artiste"
		case "label":
			artistReleasesResponse.Releases[i].TypeDisp = "Label"
		}
	}

	return artistReleasesResponse, nil
}

func getArtistDetails(artistID string) (ArtistDetail, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/artists/%s", apiURL, artistID)
	fmt.Println("Requête URL:", url) // Log de l'URL

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ArtistDetail{}, err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return ArtistDetail{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArtistDetail{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return ArtistDetail{}, fmt.Errorf("Error: %v", resp.Status)
	}

	var artistDetail ArtistDetail
	if err := json.Unmarshal(body, &artistDetail); err != nil {
		return ArtistDetail{}, err
	}

	return artistDetail, nil
}

// Fonction pour rechercher le nom d'un artiste par son ID
func getArtistNameByID(artistID string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/artists/%s", apiURL, artistID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error: %v", resp.Status)
	}

	var artistDetail ArtistDetail
	if err := json.Unmarshal(body, &artistDetail); err != nil {
		return "", err
	}

	return artistDetail.Name, nil
}

// Fonction pour rechercher le nom d'un label par son ID
func getLabelNameByID(labelID string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/labels/%s", apiURL, labelID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", OauthCrafter())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error: %v", resp.Status)
	}

	var labelDetail struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &labelDetail); err != nil {
		return "", err
	}

	return labelDetail.Name, nil
}

// Fonction pour nettoyer la description et remplacer les identifiants par les noms
func cleanAndReplaceDescription(description string) (string, error) {
	// Remplacer les balises [a=...] par le texte après le signe =
	reAEqual := regexp.MustCompile(`\[a=([^\]]+)\]`)
	description = reAEqual.ReplaceAllString(description, "$1")

	// Remplacer les balises [aXXXXXX] par les noms d'artistes
	reAID := regexp.MustCompile(`\[a(\d+)\]`)
	description = reAID.ReplaceAllStringFunc(description, func(match string) string {
		artistID := reAID.FindStringSubmatch(match)[1]
		artistName, err := getArtistNameByID(artistID)
		if err != nil {
			fmt.Printf("Erreur lors de la récupération du nom pour l'ID %s: %v\n", artistID, err)
			return match // Retourner le match original si l'artiste n'est pas trouvé
		}
		return artistName
	})

	// Remplacer les balises [l=...] par le texte après le signe =
	reLEqual := regexp.MustCompile(`\[l=([^\]]+)\]`)
	description = reLEqual.ReplaceAllString(description, "$1")

	// Remplacer les balises [lXXXXXX] par les noms de labels
	reLID := regexp.MustCompile(`\[l(\d+)\]`)
	description = reLID.ReplaceAllStringFunc(description, func(match string) string {
		labelID := reLID.FindStringSubmatch(match)[1]
		labelName, err := getLabelNameByID(labelID)
		if err != nil {
			fmt.Printf("Erreur lors de la récupération du nom pour l'ID %s: %v\n", labelID, err)
			return match // Retourner le match original si le label n'est pas trouvé
		}
		return labelName
	})

	// Supprimer les balises [b] et [/b]
	description = strings.ReplaceAll(description, "[b]", "")
	description = strings.ReplaceAll(description, "[/b]", "")

	return description, nil
}

func main() {
	// Charger d'abord les paramètres
	getSettings()

	// Charger les favoris au démarrage
	if err := loadFavorites(); err != nil {
		fmt.Println("Erreur lors du chargement des favoris:", err)
	}

	// Ensuite faire la requête OAuth
	err := requestOauthToken()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	oauthLink = fmt.Sprintf("https://discogs.com/oauth/authorize?oauth_token=%s", oauth_token)

	passedArgs := PassedVars{
		OAuthLink:  oauthLink,
		OAuthPopup: !authentified,
	}

	// Chargement des templates
	temp = template.New("")

	// Ajout des fonctions pour les templates
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	temp = temp.Funcs(funcMap)

	var errTemp error
	temp, errTemp = temp.ParseGlob("../assets/templates/*.html")
	if errTemp != nil {
		fmt.Printf("Error: %v\n", errTemp)
		return
	}

	http.HandleFunc("/authorized", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		if queryParams.Get("denied") != "" {
			fmt.Println("Accès refusé")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		oauth_verifier = queryParams.Get("oauth_verifier")
		fmt.Println(oauth_verifier)
		err := accessOauthToken()
		if err != nil {
			fmt.Println("Error [oauth_token]:", err)
		}
		if isAuthorized() {
			fmt.Println("Authentification réussie !")
		} else {
			fmt.Println("Problème d'authentification...")
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		passedArgs.OAuthPopup = !authentified
		err := "e"
		if err != "e" {
			fmt.Println("Erreur lors de la récupération des données:", err)
			http.Redirect(w, r, "/error/", http.StatusSeeOther)
		} else {
			temp.ExecuteTemplate(w, "searchBar", passedArgs)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		searchQuery := ""
		year := ""
		Type := ""
		genre := ""
		page := 1
		perPage := 10

		searchQuery = queryParams.Get("q")
		year = queryParams.Get("year")
		Type = queryParams.Get("type")
		genre = queryParams.Get("genre")

		// Récupération du numéro de page et du nombre d'éléments par page
		pageStr := queryParams.Get("page")
		if pageStr != "" {
			pageNum, err := strconv.Atoi(pageStr)
			if err == nil && pageNum > 0 {
				page = pageNum
			}
		}

		perPageStr := queryParams.Get("per_page")
		if perPageStr != "" {
			perPageNum, err := strconv.Atoi(perPageStr)
			if err == nil && perPageNum > 0 {
				perPage = perPageNum
			}
		}

		fmt.Printf("Recherche effectuée : %s (page %d, %d éléments par page)\n", searchQuery, page, perPage)

		// searchAPI(query, year, genre, type, page, perPage)
		results, err := searchAPI(searchQuery, year, genre, Type, page, perPage)
		if err != nil {
			fmt.Println("Erreur lors de la recherche:", err)
			http.Redirect(w, r, "/error/", http.StatusSeeOther)
			return
		}

		var searchResult SearchResultPvars
		searchResult.OAuthLink = oauthLink
		searchResult.OAuthPopup = !authentified
		err = json.Unmarshal(results, &searchResult)
		if err != nil {
			fmt.Println("Erreur lors du parsing des résultats:", err)
			http.Redirect(w, r, "/error/", http.StatusSeeOther)
			return
		}

		// Ajout des informations de pagination
		searchResult.CurrentPage = page
		searchResult.PerPage = perPage
		searchResult.SearchQuery = searchQuery
		searchResult.Year = year
		searchResult.Type = Type
		searchResult.Genre = genre

		// Déterminer s'il y a une page suivante
		// Si le nombre de résultats est inférieur au nombre d'éléments par page, il n'y a pas de page suivante
		searchResult.HasNextPage = len(searchResult.Results) >= perPage
		fmt.Printf("Nombre de résultats: %d\n", len(searchResult.Results))
		fmt.Printf("Nombre de résultats par page: %d\n", perPage)

		// Vérifier si chaque résultat est dans les favoris
		for i := range searchResult.Results {
			// TypeDisp est le type de la release traduit en français
			switch searchResult.Results[i].Type {
			case "release":
				searchResult.Results[i].TypeDisp = "Sortie"
			case "master":
				searchResult.Results[i].TypeDisp = "Master"
			case "artist":
				searchResult.Results[i].TypeDisp = "Artiste"
			case "label":
				searchResult.Results[i].TypeDisp = "Label"
			}

			// Remplacer le lien de l'image si c'est spacer.gif
			if strings.HasSuffix(searchResult.Results[i].ImageLink, "spacer.gif") {
				searchResult.Results[i].ImageLink = "/pictures/noPicture.png"
			}

			releaseID := strconv.Itoa(searchResult.Results[i].ID)
			if _, exists := favorites[releaseID]; exists {
				searchResult.Results[i].IsFavorite = true
			}
		}

		temp.ExecuteTemplate(w, "searchResults", searchResult)
	})

	http.HandleFunc("/testAuth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Test d'authentification")
		if isAuthorized() {
			fmt.Println("Vous êtes autorisé")
		} else {
			fmt.Println("Vous n'êtes pas autorisé")
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/favorites", func(w http.ResponseWriter, r *http.Request) {
		var favoritesPagePvars FavoritesPagePvars
		favoritesPagePvars.OAuthPopup = !authentified
		favoritesPagePvars.OAuthLink = oauthLink

		// Créer une liste temporaire pour stocker les favoris avec les types traduits
		var favoritesList []Result

		for _, favorite := range favorites {
			fmt.Println("Name:", favorite.Title, "ID:", favorite.ID, "Type:", favorite.Type)

			// Créer une copie du favori pour pouvoir la modifier
			favoriteCopy := favorite

			// Traduire le type en français
			switch favoriteCopy.Type {
			case "release":
				favoriteCopy.TypeDisp = "Sortie"
			case "master":
				favoriteCopy.TypeDisp = "Master"
			case "artist":
				favoriteCopy.TypeDisp = "Artiste"
			case "label":
				favoriteCopy.TypeDisp = "Label"
			}

			// Remplacer le lien de l'image si c'est spacer.gif
			if strings.HasSuffix(favoriteCopy.ImageLink, "spacer.gif") {
				favoriteCopy.ImageLink = "/pictures/noPicture.png"
			}

			favoritesList = append(favoritesList, favoriteCopy)
		}

		favoritesPagePvars.Favorites = favoritesList
		temp.ExecuteTemplate(w, "favoritesPage", favoritesPagePvars)
	})

	http.HandleFunc("/deniedAuth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Une erreur est survenue")
		return
	})

	http.HandleFunc("/toggleFavorite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fmt.Println("Méthode non autorisée:", r.Method)
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		var requestData struct {
			ReleaseID string `json:"releaseId"`
			Title     string `json:"title"`
			ImageLink string `json:"imageLink"`
			Country   string `json:"country"`
			Year      string `json:"year"`
			Artist    string `json:"artist"`
			Type      string `json:"type"`
			TypeDisp  string `json:"typeDisp"`
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Erreur lecture body:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &requestData); err != nil {
			fmt.Println("Erreur parsing JSON:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, _ := strconv.Atoi(requestData.ReleaseID)
		isFav := false

		// Si le favori existe, on le supprime. Sinon, on l'ajoute
		if _, exists := favorites[requestData.ReleaseID]; exists {
			delete(favorites, requestData.ReleaseID)
			fmt.Printf("Release %s retirée des favoris\n", requestData.ReleaseID)
		} else {
			favorites[requestData.ReleaseID] = Result{
				ID:         id,
				Type:       requestData.Type,
				Title:      requestData.Title,
				ImageLink:  requestData.ImageLink,
				Country:    requestData.Country,
				Year:       requestData.Year,
				TypeDisp:   requestData.TypeDisp,
				IsFavorite: true,
			}
			isFav = true
			fmt.Printf("Release %s ajoutée aux favoris\n", requestData.ReleaseID)
		}

		// Sauvegarder les changements dans le fichier
		if err := saveFavorites(); err != nil {
			fmt.Println("Erreur lors de la sauvegarde des favoris:", err)
		}

		response := struct {
			IsFavorite bool `json:"IsFavorite"`
		}{
			IsFavorite: isFav,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Println("Erreur encodage réponse:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/getFavorites", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(favorites); err != nil {
			fmt.Println("Erreur lors de l'envoi des favoris:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		releaseID := r.URL.Query().Get("id")
		if releaseID == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		releaseData, err := getReleaseDetails(releaseID)
		if err != nil {
			fmt.Println("Erreur lors de la récupération des détails de la release:", err)
			http.Redirect(w, r, "/error/", http.StatusSeeOther)
			return
		}

		var releaseDetail ReleaseDetail
		err = json.Unmarshal(releaseData, &releaseDetail)
		if err != nil {
			fmt.Println("Erreur lors du parsing des détails de la release:", err)
			http.Redirect(w, r, "/error/", http.StatusSeeOther)
			return
		}

		// Vérifier si la release est dans les favoris
		releaseIDStr := strconv.Itoa(releaseDetail.ID)
		if _, exists := favorites[releaseIDStr]; exists {
			// Ajouter un indicateur pour afficher l'étoile remplie
		}

		releaseDetail.OAuthLink = oauthLink
		releaseDetail.OAuthPopup = !authentified

		temp.ExecuteTemplate(w, "releaseDisplay", releaseDetail)
	})

	http.HandleFunc("/master", func(w http.ResponseWriter, r *http.Request) {
		masterID := r.URL.Query().Get("id")
		if masterID == "" {
			http.Error(w, "ID du master non spécifié", http.StatusBadRequest)
			return
		}

		// Récupérer les détails du master
		masterDetail, err := getMasterDetails(masterID)

		masterDetail.OAuthLink = oauthLink
		masterDetail.OAuthPopup = !authentified

		if err != nil {
			http.Error(w, "Erreur lors de la récupération des détails du master: "+err.Error(), http.StatusInternalServerError)
			return
		}

		temp.ExecuteTemplate(w, "masterDisplay", masterDetail)
	})

	http.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
		// Extraire l'ID de l'artiste de l'URL
		artistID := r.URL.Query().Get("id")
		if artistID == "" {
			http.Error(w, "ID de l'artiste non spécifié", http.StatusBadRequest)
			return
		}

		// Récupérer les détails de l'artiste
		artistDetail, err := getArtistDetails(artistID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des détails de l'artiste: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Nettoyer et remplacer les identifiants par les noms
		artistDetail.Profile, err = cleanAndReplaceDescription(artistDetail.Profile)
		if err != nil {
			fmt.Println("Erreur lors du nettoyage de la description:", err)
			return
		}

		artistDetail.OAuthLink = oauthLink
		artistDetail.OAuthPopup = !authentified
		temp.ExecuteTemplate(w, "artistDisplay", artistDetail)
	})

	http.HandleFunc("/artistReleases/", func(w http.ResponseWriter, r *http.Request) {
		// Extraire l'ID de l'artiste de l'URL
		artistID := r.URL.Query().Get("id")
		page := r.URL.Query().Get("page")
		perPage := r.URL.Query().Get("per_page")
		releaseType := r.URL.Query().Get("type")
		year := r.URL.Query().Get("year")

		if page == "" {
			page = "1"
		}
		if perPage == "" {
			perPage = "10"
		}

		if artistID == "" {
			http.Error(w, "ID de l'artiste non spécifié", http.StatusBadRequest)
			return
		}

		pageInt, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, "Numéro de page invalide", http.StatusBadRequest)
			return
		}

		perPageInt, err := strconv.Atoi(perPage)
		if err != nil {
			http.Error(w, "Nombre d'éléments par page invalide", http.StatusBadRequest)
			return
		}

		artistReleasesResponse, err := getArtistReleases(artistID, pageInt, perPageInt, releaseType, year)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des releases de l'artiste: "+err.Error(), http.StatusInternalServerError)
			return
		}

		artistReleasesResponse.OAuthLink = oauthLink
		artistReleasesResponse.OAuthPopup = !authentified
		artistReleasesResponse.ArtistID = artistID
		artistReleasesResponse.ArtistName, err = getArtistNameByID(artistID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération du nom de l'artiste: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Vérifier si chaque release est dans les favoris
		for i := range artistReleasesResponse.Releases {
			releaseID := strconv.Itoa(artistReleasesResponse.Releases[i].ID)
			if _, exists := favorites[releaseID]; exists {
				artistReleasesResponse.Releases[i].IsFavorite = true
			}
		}

		temp.ExecuteTemplate(w, "artistReleases", artistReleasesResponse)
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/LDelaforet/Groupie-Tracker/blob/master/README.md", http.StatusSeeOther)
	})

	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("../assets/templates"))))
	http.Handle("/pictures/", http.StripPrefix("/pictures/", http.FileServer(http.Dir("../assets/pictures/"))))
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("../assets/styles/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("../assets/fonts/"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("../assets/scripts/"))))

	fmt.Println("Serveur lancé sur", host+":"+strconv.Itoa(port))
	http.ListenAndServe(host+":"+strconv.Itoa(port), nil)
}
