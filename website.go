package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db       *sql.DB
	dlsource = "http://q2.pfl.gr/dl/baseq2/maps"
)

type MapData struct {
	Name        string
	Description string
	Textures    []string
	Download    string
	Year        int
	Weapons     []template.HTML
	Ammo        []template.HTML
	Items       []template.HTML
}

type IndexData struct {
	Feature MapData
	Cards   []MapData
	Year    int
}

func main() {
	var (
		host         = flag.String("h", "[::]", "IP to listen on")
		port         = flag.Int("p", 2222, "Webserver listen port")
		databasefile = flag.String("db", "mapdata.sqlite", "The SQLITE database file")
	)
	flag.Parse()

	database, err := sql.Open("sqlite3", *databasefile)
	if err != nil {
		panic(err)
	}
	db = database

	r := mux.NewRouter()
	r.HandleFunc("/map/{map}", WebMapView)
	r.HandleFunc("/list", WebList)
	r.HandleFunc("/search", WebSearch)
	r.HandleFunc("/", WebIndex)
	r.PathPrefix("/static").Handler(loggingHandler(http.FileServer(http.Dir("./website"))))

	httpsrv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%d", *host, *port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening for web requests on %s\n", httpsrv.Addr)
	log.Fatal(httpsrv.ListenAndServe())
}

// for logging built-in http.FileServer requests
func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		h.ServeHTTP(w, r)
	})
}

// Show a main map at the top and 3 little ones at the bottom
func WebIndex(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	minis := []MapData{}
	feat := FrontPageMap()
	for i := 0; i < 3; i++ {
		minis = append(minis, FrontPageMap())
	}
	data := IndexData{
		Feature: feat,
		Cards:   minis,
		Year:    time.Now().Year(),
	}

	t, err := template.ParseFiles("website/template/index.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = t.ExecuteTemplate(w, "main", data)
	if err != nil {
		log.Println(err)
	}
}

func WebSearch(w http.ResponseWriter, r *http.Request) {
	log.Println("search view")
}

func WebList(w http.ResponseWriter, r *http.Request) {
}

func WebMapView(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	weaps := []template.HTML{}
	ammo := []template.HTML{}
	it := []template.HTML{}
	vars := mux.Vars(r)
	data := MapData{
		Name: vars["map"],
	}
	data.Textures = GetMapTextures(data.Name)
	data.Download = dlsource + "/" + data.Name + ".bsp"
	data.Year = time.Now().Year()

	weapons, ammos, items := GetEntityInfo(data.Name)
	for k, v := range weapons {
		weaps = append(weaps, template.HTML(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", k, v)))
	}
	for k, v := range ammos {
		ammo = append(ammo, template.HTML(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", k, v)))
	}
	for k, v := range items {
		it = append(it, template.HTML(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", k, v)))
	}

	data.Weapons = weaps
	data.Ammo = ammo
	data.Items = it

	t, err := template.ParseFiles(
		"website/template/index.tmpl",
		"website/template/view.tmpl",
	)
	if err != nil {
		log.Println(err)
		return
	}
	err = t.ExecuteTemplate(w, "viewmap", data)
	if err != nil {
		log.Println(err)
	}
}

func FrontPageMap() MapData {
	var mapname, descr string
	s := "SELECT name, '' FROM map ORDER BY RANDOM() LIMIT 1"
	err := db.QueryRow(s).Scan(&mapname, &descr)
	if err != nil {
		log.Println(err)
		return MapData{Name: "unknown", Description: "unknown"}
	}
	return MapData{Name: mapname, Description: descr}
}

// Grab all the textures for a specific map from the database
func GetMapTextures(mapname string) []string {
	textures := []string{}
	t := ""
	s := `	SELECT texture 
			FROM texture t
			JOIN map m ON m.id = t.map 
			WHERE m.name = ?`
	rows, err := db.Query(s, mapname)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer rows.Close()
	for rows.Next() {
		e := rows.Scan(&t)
		if e != nil {
			log.Println(e)
			continue
		}
		textures = append(textures, strings.Trim(t, "�"))
	}
	return textures
}

// Add the request to the log
func logRequest(r *http.Request) {
	log.Println(r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
}

func GetEntityInfo(mapname string) (map[string]int, map[string]int, map[string]int) {
	weaps := make(map[string]int)
	ammo := make(map[string]int)
	items := make(map[string]int)
	s := `	SELECT classname, quantity 
			FROM entity e
			JOIN map m ON m.id = e.map 
			WHERE m.name = ?
			ORDER BY classname`
	rows, err := db.Query(s, mapname)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	item := ""
	qty := 0
	for rows.Next() {
		err := rows.Scan(&item, &qty)
		if err != nil {
			log.Println(err)
			continue
		}
		switch item {
		case "weapon_shotgun":
			weaps["shotgun"] = qty
		case "weapon_supershotgun":
			weaps["super shotgun"] = qty
		case "weapon_machinegun":
			weaps["machinegun"] = qty
		case "weapon_chaingun":
			weaps["chaingun"] = qty
		case "weapon_grenadelauncher":
			weaps["grenade launcher"] = qty
		case "weapon_hyperblaster":
			weaps["hyperblaster"] = qty
		case "weapon_rocketlauncher":
			weaps["rocket launcher"] = qty
		case "weapon_railgun":
			weaps["railgun"] = qty
		case "weapon_bgf":
			weaps["bfg"] = qty
		case "item_adrenaline":
			items["adrenaline"] = qty
		case "item_ancient_head":
			items["ancient head"] = qty
		case "item_armor_combat":
			items["armor - yellow"] = qty
		case "item_armor_body":
			items["armor - red"] = qty
		case "item_armor_jacket":
			items["armor - green"] = qty
		case "item_armor_shard":
			items["shards"] = qty
		case "item_bandolier":
			items["bandolier"] = qty
		case "item_health":
			items["health - small"] = qty
		case "item_health_large":
			items["health - large"] = qty
		case "item_health_mega":
			items["megahealth"] = qty
		case "item_invulnerability":
			items["invuln"] = qty
		case "item_pack":
			items["backpack"] = qty
		case "item_power_screen":
			items["power screen"] = qty
		case "item_power_shield":
			items["power shield"] = qty
		case "item_quad":
			items["quad damage"] = qty
		case "item_silencer":
			items["silencer"] = qty
		case "ammo_shells":
			ammo["shells"] = qty
		case "ammo_bullets":
			ammo["bullets"] = qty
		case "ammo_grenades":
			ammo["grenades"] = qty
		case "ammo_cells":
			ammo["cells"] = qty
		case "ammo_rockets":
			ammo["rockets"] = qty
		case "ammo_slugs":
			ammo["slugs"] = qty
		}

	}
	return weaps, ammo, items
}
