package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/spf13/viper"
    "gopkg.in/pg.v4"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"
)

var PLATFORM_IDS = map[string]string{
    "BR": "BR1",
    "EUNE": "EUN1",
    "EUW": "EUW1",
    "KR": "KR",
    "LAN": "LA1",
    "LAS": "LA2",
    "NA": "NA1",
    "OCE": "OC1",
    "TR": "TR1",
    "RU": "RU",
}
const LOL_API_HOST_SUFFIX = "api.pvp.net"
var CHAMPIONS = [...]string { "shaco","drmundo","rammus","anivia","irelia","yasuo","sona","kassadin","zac","gnar","karma","corki","gangplank","janna","jhin","kindred","braum","ashe","tryndamere","jax","morgana","zilean","singed","evelynn","twitch","galio","velkoz","olaf","annie","karthus","leblanc","urgot","amumu","xinzhao","chogath","twistedfate","fiddlesticks","vladimir","warwick","teemo","tristana","sivir","soraka","ryze","sion","masteryi","alistar","missfortune","nunu","rengar","volibear","fizz","graves","ahri","shyvana","lux","xerath","thresh","shen","kogmaw","jinx","tahmkench","riven","talon","malzahar","kayle","kalista","reksai","illaoi","leona","lulu","gragas","poppy","fiora","ziggs","udyr","viktor","sejuani","varus","nautilus","draven","bard","mordekaiser","ekko","yorick","pantheon","ezreal","garen","akali","kennen","vayne","jayce","lissandra","cassiopeia","rumble","khazix","darius","hecarim","skarner","lucian","heimerdinger","nasus","zed","nidalee","syndra","jarvaniv","quinn","renekton","maokai","aurelionsol","nocturne","katarina","leesin","monkeyking","azir","brand","diana","elise","nami","aatrox","orianna","zyra","trundle","veigar","taric","caitlyn","blitzcrank","malphite","vi","swain" }
var CHAMPION_KEYS = map[string]string{ "aatrox":"266","ahri":"103","akali":"84","alistar":"12","amumu":"32","anivia":"34","annie":"1","ashe":"22","aurelionsol":"136","azir":"268","bard":"432","blitzcrank":"53","brand":"63","braum":"201","caitlyn":"51","cassiopeia":"69","chogath":"31","corki":"42","darius":"122","diana":"131","draven":"119","drmundo":"36","ekko":"245","elise":"60","evelynn":"28","ezreal":"81","fiddlesticks":"9","fiora":"114","fizz":"105","galio":"3","gangplank":"41","garen":"86","gnar":"150","gragas":"79","graves":"104","hecarim":"120","heimerdinger":"74","illaoi":"420","irelia":"39","janna":"40","jarvaniv":"59","jax":"24","jayce":"126","jhin":"202","jinx":"222","kalista":"429","karma":"43","karthus":"30","kassadin":"38","katarina":"55","kayle":"10","kennen":"85","khazix":"121","kindred":"203","kogmaw":"96","leblanc":"7","leesin":"64","leona":"89","lissandra":"127","lucian":"236","lulu":"117","lux":"99","malphite":"54","malzahar":"90","maokai":"57","masteryi":"11","missfortune":"21","monkeyking":"62","wukong":"62","mordekaiser":"82","morgana":"25","nami":"267","nasus":"75","nautilus":"111","nidalee":"76","nocturne":"56","nunu":"20","olaf":"2","orianna":"61","pantheon":"80","poppy":"78","quinn":"133","rammus":"33","reksai":"421","renekton":"58","rengar":"107","riven":"92","rumble":"68","ryze":"13","sejuani":"113","shaco":"35","shen":"98","shyvana":"102","singed":"27","sion":"14","sivir":"15","skarner":"72","sona":"37","soraka":"16","swain":"50","syndra":"134","tahmkench":"223","talon":"91","taric":"44","teemo":"17","thresh":"412","tristana":"18","trundle":"48","tryndamere":"23","twistedfate":"4","twitch":"29","udyr":"77","urgot":"6","varus":"110","vayne":"67","veigar":"45","velkoz":"161","vi":"254","viktor":"112","vladimir":"8","volibear":"106","warwick":"19","xerath":"101","xinzhao":"5","yasuo":"157","yorick":"83","zac":"154","zed":"238","ziggs":"115","zilean":"26","zyra":"143"}
var CHAMPION_KEYS_BY_KEY = map[string]string{ "1":"annie", "2":"olaf", "3":"galio", "4":"twistedfate", "5":"xinzhao", "6":"urgot", "7":"leblanc", "8":"vladimir", "9":"fiddlesticks", "10":"kayle", "11":"masteryi", "12":"alistar", "13":"ryze", "14":"sion", "15":"sivir", "16":"soraka", "17":"teemo", "18":"tristana", "19":"warwick", "20":"nunu", "21":"missfortune", "22":"ashe", "23":"tryndamere", "24":"jax", "25":"morgana", "26":"zilean", "27":"singed", "28":"evelynn", "29":"twitch", "30":"karthus", "31":"chogath", "32":"amumu", "33":"rammus", "34":"anivia", "35":"shaco", "36":"drmundo", "37":"sona", "38":"kassadin", "39":"irelia", "40":"janna", "41":"gangplank", "42":"corki", "43":"karma", "44":"taric", "45":"veigar", "48":"trundle", "50":"swain", "51":"caitlyn", "53":"blitzcrank", "54":"malphite", "55":"katarina", "56":"nocturne", "57":"maokai", "58":"renekton", "59":"jarvaniv", "60":"elise", "61":"orianna", "62":"monkeyking", "63":"brand", "64":"leesin", "67":"vayne", "68":"rumble", "69":"cassiopeia", "72":"skarner", "74":"heimerdinger", "75":"nasus", "76":"nidalee", "77":"udyr", "78":"poppy", "79":"gragas", "80":"pantheon", "81":"ezreal", "82":"mordekaiser", "83":"yorick", "84":"akali", "85":"kennen", "86":"garen", "89":"leona", "90":"malzahar", "91":"talon", "92":"riven", "96":"kogmaw", "98":"shen", "99":"lux", "101":"xerath", "102":"shyvana", "103":"ahri", "104":"graves", "105":"fizz", "106":"volibear", "107":"rengar", "110":"varus", "111":"nautilus", "112":"viktor", "113":"sejuani", "114":"fiora", "115":"ziggs", "117":"lulu", "119":"draven", "120":"hecarim", "121":"khazix", "122":"darius", "126":"jayce", "127":"lissandra", "131":"diana", "133":"quinn", "134":"syndra", "136":"aurelionsol", "143":"zyra", "150":"gnar", "154":"zac", "157":"yasuo", "161":"velkoz", "201":"braum", "202":"jhin", "203":"kindred", "222":"jinx", "223":"tahmkench", "236":"lucian", "238":"zed", "245":"ekko", "254":"vi", "266":"aatrox", "267":"nami", "268":"azir", "412":"thresh", "420":"illaoi", "421":"reksai", "429":"kalista", "432":"bard"}
var CHAMPION_KEYS_BY_KEY_PROPER_CASING = map[string]string{ "35": "Shaco", "36": "DrMundo", "33": "Rammus", "34": "Anivia", "39": "Irelia", "157": "Yasuo", "37": "Sona", "38": "Kassadin", "154": "Zac", "150": "Gnar", "43": "Karma", "42": "Corki", "41": "Gangplank", "40": "Janna", "202": "Jhin", "203": "Kindred", "201": "Braum", "22": "Ashe", "23": "Tryndamere", "24": "Jax", "25": "Morgana", "26": "Zilean", "27": "Singed", "28": "Evelynn", "29": "Twitch", "3": "Galio", "161": "Velkoz", "2": "Olaf", "1": "Annie", "30": "Karthus", "7": "Leblanc", "6": "Urgot", "32": "Amumu", "5": "XinZhao", "31": "Chogath", "4": "TwistedFate", "9": "FiddleSticks", "8": "Vladimir", "19": "Warwick", "17": "Teemo", "18": "Tristana", "15": "Sivir", "16": "Soraka", "13": "Ryze", "14": "Sion", "11": "MasterYi", "12": "Alistar", "21": "MissFortune", "20": "Nunu", "107": "Rengar", "106": "Volibear", "105": "Fizz", "104": "Graves", "103": "Ahri", "102": "Shyvana", "99": "Lux", "101": "Xerath", "412": "Thresh", "98": "Shen", "96": "KogMaw", "222": "Jinx", "223": "TahmKench", "92": "Riven", "91": "Talon", "90": "Malzahar", "10": "Kayle", "429": "Kalista", "421": "RekSai", "420": "Illaoi", "89": "Leona", "117": "Lulu", "79": "Gragas", "78": "Poppy", "114": "Fiora", "115": "Ziggs", "77": "Udyr", "112": "Viktor", "113": "Sejuani", "110": "Varus", "111": "Nautilus", "119": "Draven", "432": "Bard", "82": "Mordekaiser", "245": "Ekko", "83": "Yorick", "80": "Pantheon", "81": "Ezreal", "86": "Garen", "84": "Akali", "85": "Kennen", "67": "Vayne", "126": "Jayce", "127": "Lissandra", "69": "Cassiopeia", "68": "Rumble", "121": "Khazix", "122": "Darius", "120": "Hecarim", "72": "Skarner", "236": "Lucian", "74": "Heimerdinger", "75": "Nasus", "238": "Zed", "76": "Nidalee", "134": "Syndra", "59": "JarvanIV", "133": "Quinn", "58": "Renekton", "57": "Maokai", "136": "AurelionSol", "56": "Nocturne", "55": "Katarina", "64": "LeeSin", "62": "MonkeyKing", "268": "Azir", "63": "Brand", "131": "Diana", "60": "Elise", "267": "Nami", "266": "Aatrox", "61": "Orianna", "143": "Zyra", "48": "Trundle", "45": "Veigar", "44": "Taric", "51": "Caitlyn", "53": "Blitzcrank", "54": "Malphite", "254": "Vi", "50": "Swain"}

type Mastery struct {
    ChampionId                   int   `json:"championId"`
    ChampionLevel                int   `json:"championLevel"`
    ChampionPoints               int   `json:"championPoints"`
    ChampionPointsSinceLastLevel int64 `json:"championPointsSinceLastLevel"`
    ChampionPointsUntilNextLevel int64 `json:"championPointsUntilNextLevel"`
    LastPlayTime                 int64 `json:"lastPlayTime"`
    SummonerId                   int64 `json:"playerId"`
}

type RiotError struct {
    StatusCode int
}

func (err RiotError) Error() string {
    return fmt.Sprintf("Error: HTTP Status %d", err.StatusCode)
}

type Summoner struct {
    SummonerId    int64  `json:"id"`
    Name          string `json:"name"`
    ProfileIconID int    `json:"profileIconId"`
    MasteriesUpdatedAt time.Time
    RevisionDate  int64  `json:"revisionDate"`
    SummonerLevel int    `json:"summonerLevel"`
}

func (s *Summoner) SetSummoner(name string) {
    s.Name = name
}

// type Champions struct {
//     Page  int `json:"page"`
//     Limit int `json:"limit"`
//     Data  []ChampionRoleData
// }

// type ChampionRoleData struct {
//     Key  string `json:"key"`
//     Role string `json:"role"`
//     Name string `json:"name"`
// }

type ChampionMatchupWinrate struct {
    Role string `json:"role"`
    Matchups []Matchup
}

type Matchup struct {
    Games int `json:"games"`
    StatScore float32 `json:"statScore"`
    WinRate float32 `json:"winRate"`
    Enemy string `json:"key"`
}

type ChampionMatchup struct {
    Champion string
    Enemy string
    Games int
    Role string
    StatScore float32
    WinRate float32
}

func (c *ChampionMatchup) SetChampion(name string) {
    c.Champion = name
}

func (m ChampionMatchup) String() string {
    return fmt.Sprintf("[%s] %s vs %s. %g %d games\n", m.Role, CHAMPION_KEYS_BY_KEY[m.Champion], CHAMPION_KEYS_BY_KEY[m.Enemy], m.WinRate, m.Games)
}

func createSchema(db *pg.DB) error {
    queries := []string{
        // `DROP TABLE champion_matchups`,
        // `DROP TABLE masteries`,
        // `DROP TABLE summoners`,

        `CREATE TABLE IF NOT EXISTS champion_matchups (
            champion text,
            enemy text,
            games int,
            role text,
            stat_score decimal,
            win_rate decimal,
            PRIMARY KEY(champion, enemy, role))`,

        `CREATE TABLE IF NOT EXISTS masteries (
            champion_id bigint,
            champion_level int,
            champion_points int,
            champion_points_since_last_level bigint,
            champion_points_until_next_level bigint,
            last_play_time bigint,
            summoner_id bigint,
            PRIMARY KEY (champion_id, summoner_id))`,

        `CREATE TABLE IF NOT EXISTS summoners (
            name text,
            profile_icon_id int,
            masteries_updated_at timestamp with time zone DEFAULT (now() at time zone 'utc'),
            revision_date bigint,
            summoner_level int,
            summoner_id bigint PRIMARY KEY)`,

        `CREATE OR REPLACE FUNCTION upsert_masteries(c_id bigint, c_lv int, c_pts int, c_pts_since_last_lv bigint, c_pts_until_next_lv bigint, last_play_t bigint, s_id bigint) RETURNS VOID AS $$
            DECLARE
            BEGIN
                UPDATE masteries SET champion_id = c_id, champion_level = c_lv, champion_points = c_pts, champion_points_since_last_level = c_pts_since_last_lv, champion_points_until_next_level = c_pts_until_next_lv, last_play_time = last_play_t, summoner_id = s_id WHERE champion_id = c_id AND summoner_id = s_id;
                IF NOT FOUND THEN
                INSERT INTO masteries VALUES (c_id, c_lv, c_pts, c_pts_since_last_lv, c_pts_until_next_lv, last_play_t, s_id);
                END IF;
            END;
            $$ LANGUAGE 'plpgsql'`,

    }

    for _, q := range queries {
        _, err := db.Exec(q)
        if err != nil {
            fmt.Println(q)
            return err
        }
    }
    return nil
}

func getSummonerMasteriesAndSave(summonerName string, db *pg.DB) (err error) {
    name := NormalizeSummonerName(summonerName)[0]
    summoners, err := getSummonerIdByNameAndSave("NA", name, db)
    if err != nil {
        fmt.Println(err)
        return
    }
    _, err = getChampionMasteriesBySummonerIdAndSave("NA", summoners[name].SummonerId, db)
    return
}

func getSummonerIdByNameAndSave(region string, name string, db *pg.DB) (summoners map[string]Summoner, err error) {
    args := "api_key=" + viper.GetString("riot.key")
    // names := strURLParameter(name).String()
    summoners = make(map[string]Summoner)
    url := fmt.Sprintf(
            "https://%v.%v/api/lol/%v/v1.4/summoner/by-name/%v?%v",
            region,
            LOL_API_HOST_SUFFIX,
            region,
            name,
            args)
    fmt.Println(url)
    err = requestAndUnmarshal(url, &summoners)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(summoners[name])
    if summoners[name].SummonerId == 0 {
        fmt.Println(err)
        return
    }
    s := summoners[name]
    s.SetSummoner(name)
    s.MasteriesUpdatedAt = time.Now().UTC()
    fmt.Println(s)
    err = db.Create(&s)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(summoners[name].SummonerId)
    return
}

// func getAllChampions(region string) (champions, err error) {
//     /api/lol/static-data/{region}/v1.2/champion
// }

func getChampionMasteriesBySummonerIdAndSave(region string, summonerId int64, db *pg.DB) (masteries []Mastery, err error) {
    args := "api_key=" + viper.GetString("riot.key")
    platformId := PLATFORM_IDS[region]
    // var masteries []Mastery
    url := fmt.Sprintf(
            "https://%v.%v/championmastery/location/%v/player/%v/champions?%v",
            region,
            LOL_API_HOST_SUFFIX,
            platformId,
            summonerId,
            args)
    fmt.Println(url)
    err = requestAndUnmarshal(url, &masteries)
    if err != nil {
        fmt.Println(err)
        return
    }
    // var savedMasteries []Mastery
    // err = db.Model(&masteries).Where("summoner_id = ?", summonerId).Select()

    // if err != nil {
    //     // if err.Error() == "pg: no rows in result set" {
    //     //     savedMasteries = nil
    //     // } else {
    //         fmt.Println(err)
    //         return
    //     // }
    // }

    for _, mastery := range masteries {
        sql := fmt.Sprintf(
            "SELECT upsert_masteries(%v, %v, %v, %v, %v, %v, %v)",
            mastery.ChampionId,
            mastery.ChampionLevel,
            mastery.ChampionPoints,
            mastery.ChampionPointsSinceLastLevel,
            mastery.ChampionPointsUntilNextLevel,
            mastery.LastPlayTime,
            mastery.SummonerId)
        _, err = db.Exec(sql)
        // err = db.Create(&mastery)
        if err != nil {
            fmt.Println(sql)
            fmt.Println(err)
            return
        }
    }
    // fmt.Println(masteries)
    return
}

func getWinrateForChampion(champion string, db *pg.DB) (matchups []ChampionMatchup, err error) {
    var matchupInfo []ChampionMatchupWinrate
    url := fmt.Sprintf(
        "http://api.champion.gg/champion/%v/matchup?api_key=%v",
        champion,
        viper.GetString("championgg.key"))
    fmt.Println(url)
    err = requestAndUnmarshal(url, &matchupInfo)
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, data := range matchupInfo {
        role := data.Role
        for _, matchup := range data.Matchups {
            c := ChampionMatchup{
                Champion: CHAMPION_KEYS[champion],
                Enemy: CHAMPION_KEYS[NormalizeChampion(matchup.Enemy)],
                Games: matchup.Games,
                Role: role,
                StatScore: matchup.StatScore,
                WinRate: matchup.WinRate}
            err = db.Create(&c)
            if err != nil {
                fmt.Println(err)
                return
            }
            matchups = append(matchups, c)
        }
    }
    // fmt.Println(matchups)
    return
}

/* ======================== db methods ======================= */

func getOrCreateSummoner(summonerName string, db *pg.DB) (summoner Summoner, err error) {
    name := NormalizeSummonerName(summonerName)[0]
    err = db.Model(&summoner).Where("name = ?", name).Select()
    if err != nil {
        if err.Error() == "pg: no rows in result set" {
            err = getSummonerMasteriesAndSave(name, db)
            if err != nil {
                fmt.Println(err)
                return
            }
            err = db.Model(&summoner).Where("name = ?", name).Select()
            if err != nil {
                fmt.Println(err)
                return
            }
        } else {
            fmt.Println(err)
            return
        }
    }
    // update old masteries
    // if summoner.MasteriesUpdatedAt.UTC().Before(time.Now().UTC().Add(time.Duration(60*60*24)*time.Second)) {
    //     _, err = getChampionMasteriesBySummonerIdAndSave("NA", summoner.SummonerId, db)
    //     if err != nil {
    //         fmt.Println(err)
    //         return
    //     }
    // }

    fmt.Println("getSummonerById", name, summoner)
    return
}

func getMatchups(summoner_id int64, enemy_champion_id string, role string, db *pg.DB) (matchups []ChampionMatchup,  err error) {
    //select * from champion_matchups where enemy = '57' and champion IN (select cast(champion_id as text)  from masteries where summoner_id = 26691960) order by win_rate desc;

    // sql := fmt.Sprintf("SELECT * FROM champion_matchups WHERE enemy = '%s' AND champion IN (SELECT CAST(champion_id AS text) FROM masteries WHERE summoner_id = %d) ORDER BY win_rate DESC",
    //     enemy_champion_id,
    //     summoner_id)
    // fmt.Println(sql)
    // res, err := db.Exec(sql)

    err = db.Model(&matchups).Where("role = ? AND enemy = ? AND champion IN (SELECT  CAST(champion_id AS text) FROM masteries WHERE summoner_id = ?)", role, enemy_champion_id, summoner_id).Order("win_rate desc").Select()
    if err != nil {
        fmt.Println(err)
        return
    }
    // fmt.Println(matchups)
    return
}

/* ======================== helper ========================== */

func requestAndUnmarshal(requestURL string, v interface{}) (err error) {
    resp, err := http.Get(requestURL)
    defer resp.Body.Close()
    if err != nil {
        fmt.Println(err)
        return
    }
    if resp.StatusCode != http.StatusOK {
        return RiotError{StatusCode: resp.StatusCode}
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }

    err = json.Unmarshal(body, v)
    if err != nil {
        fmt.Println(err)
        return
    }
    return
}

func createSummonerIDString(summonerID []int64) (summonerIDstr string, err error) {
    if len(summonerID) > 40 {
        return summonerIDstr, errors.New("A Maximum of 40 SummonerIDs are allowed")
    }
    for k, v := range summonerID {
        summonerIDstr += strconv.FormatInt(v, 10)
        if k != len(summonerID)-1 {
            summonerIDstr += ","
        }
    }
    return
}

//NormalizeSummonerName takes an arbitrary number of strings and returns a string array containing the strings
//standardized to league of legends internal standard (lowecase and strings removed)
func NormalizeSummonerName(summonerNames ...string) []string {
    for i, v := range summonerNames {
        summonerName := strings.ToLower(v)
        summonerName = strings.Replace(summonerName, " ", "", -1)
        summonerNames[i] = summonerName
    }
    return summonerNames
}

func NormalizeChampion(name string) string {
    name = strings.ToLower(name)
    name = strings.Replace(name, " ", "", -1)
    name = strings.Replace(name, "'", "", -1)
    return name
}