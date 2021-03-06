package main

import (
    "errors"
    "fmt"
    "html/template"
    "github.com/spf13/viper"
    "gopkg.in/pg.v4"
    "log"
    "net/http"
    "strings"
)

var db *pg.DB
var urlPrefix string

type IndexResult struct {
    Prefix string
    Error  error
}

type MatchupResults struct {
    Enemy         string
    Prefix        string
    Role          string
    SummonerName  string
    Matchups      []ChampionMatchup
    ResultsLength int
}

func main() {

    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.SetConfigType("json")
    err := viper.ReadInConfig()
    if err != nil {
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }
    username := viper.GetString("postgres.username")
    password := viper.GetString("postgres.password")
    urlPrefix = viper.GetString("urlPrefix")

    db = pg.Connect(&pg.Options{
        User: username,
        Password: password,
    })

    err = createSchema(db)
    if err != nil {
        panic(err)
    }

    err = setupRiotApi()
    if err != nil {
        panic(err)
    }

    // var summoner MySummoner
    // err := db.Model(&summoner).Column("summoner.*", "Masteries").First()
    // if err != nil {
    //     panic(err)
    // }

    // select win_rate, enemy from champion_matchups where champion = 'Tryndamere' and role = 'Top' and games > 50 order by win_rate desc;
    // select win_rate, champion from champion_matchups where enemy = 'Garen' and role = 'Top' and games > 50 order by win_rate desc;

    // enemyChamp := "Maokai"
    // enemyKey := CHAMPION_KEYS[enemyChamp]
    // getSummonerMasteriesAndSave("Yaw Hide", db)


    // getSummonerMasteriesAndSave("EuglossaCognata", db)
    // // get all win rates between champions

    // for _, champion := range CHAMPIONS {
    //     getWinrateForChampion(champion, db)
    // }


    // summoner, err := getSummonerById(NormalizeSummonerName("Yaw Hide"), db)
    // if err != nil {
    //     panic(err)
    // }
    // getMatchups(summoner.SummonerId, "57", db)

    serveSingle("/sitemap.xml", "./static/sitemap.xml")
    serveSingle("/favicon.ico", "./static/favicon.ico")
    serveSingle("/robots.txt", "./static/robots.txt")

    http.HandleFunc(urlPrefix + "", Index)
    http.HandleFunc(urlPrefix + "matchup", GetMatchup)
    fs := http.FileServer(http.Dir("static"))
    http.Handle(urlPrefix + "static/", http.StripPrefix(urlPrefix + "static/", fs))
    fmt.Println("Server started")
    // analytics
    http.HandleFunc(urlPrefix + "analytics/index", AnalyzeIndex)
    http.HandleFunc(urlPrefix + "analytics/matchup", AnalyzeMatchup)
    http.HandleFunc(urlPrefix + "analytics/external", AnalyzeExternalLink)
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func serveSingle(pattern string, filename string) {
    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filename)
    })
}

func Index(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        result := IndexResult{urlPrefix, nil}
        t, _ := template.ParseFiles("index.html")
        a := AnalyticsPage{"/", r.Referer()}
        // fmt.Println(a)
        err := db.Create(&a)
        if err != nil {
            fmt.Println("Failed to save page analyics", err)
        }
        t.Execute(w, result)
    } else {
        http.Redirect(w, r, "/", 301)
    }
}

func GetMatchup(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // redirect to /
        http.Redirect(w, r, "/", 301)
    } else if r.Method == "POST" {
        r.ParseForm()
        enemy := r.Form.Get("enemy")
        region := strings.ToLower(r.Form.Get("region"))
        role := r.Form.Get("role")
        summonerName := r.Form.Get("name")
        if summonerName == "" || enemy == "" || role == ""  || !RIOT_REGIONS[region] {
            var err string
            if summonerName == "" {
                err = "Summoner Name"
            } else if enemy == "" {
                err = "Enemy Champion"
            } else if role == "" {
                err = "Role"
            } else if !RIOT_REGIONS[region] {
                err = "Region"
            }
            t, _ := template.ParseFiles("index.html")
            result := IndexResult{urlPrefix, errors.New(err)}
            t.Execute(w, result)
            return
        }

        enemy = NormalizeChampion(enemy)
        if CHAMPION_KEYS[enemy] == "" {
            t, _ := template.ParseFiles("index.html")
            result := IndexResult{urlPrefix, errors.New("Enemy Champion")}
            t.Execute(w, result)
            return
        }

        // fmt.Println("enemy:", enemy, CHAMPION_KEYS[enemy])
        // fmt.Println("region:", region)
        // fmt.Println("role:", role)
        // fmt.Println("summoner name:", summonerName)

        a := AnalyticsPage{"/matchup", r.Referer()}
        // fmt.Println(a)
        err := db.Create(&a)
        if err != nil {
            fmt.Println("Failed to save index analyics", err)
        }

        summoner, err := getOrCreateSummoner(region, summonerName, db)
        if err != nil {
            t, _ := template.ParseFiles("index.html")
            fmt.Println("Error get or create summoner", err)
            result := IndexResult{urlPrefix, err}
            t.Execute(w, result)
            return
        }
        matchups, pm, err := getMatchups(summoner.SummonerId, CHAMPION_KEYS[enemy], role, db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        for index, matchup := range matchups {
            matchups[index].SetChampion(CHAMPION_KEYS_BY_KEY_PROPER_CASING[matchup.Champion])
            for _, p := range pm {
                if p.Champion == matchup.Champion {
                    matchups[index].UpdatePersonalData(p)
                }
            }
        }
        // fmt.Println(CHAMPION_KEYS_BY_KEY_PROPER_CASING[CHAMPION_KEYS[enemy]], urlPrefix, role, summonerName, matchups)
        result := MatchupResults{CHAMPION_KEYS_BY_KEY_PROPER_CASING[CHAMPION_KEYS[enemy]], urlPrefix, role, summonerName, matchups, len(matchups)}
        t, _ := template.ParseFiles("matchups.html")
        t.Execute(w, result)
    } else {
        http.Redirect(w, r, "/", 301)
    }
}
