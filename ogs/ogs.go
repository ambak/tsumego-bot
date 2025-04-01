package ogs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ambak/tsumego-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
	"github.com/go-co-op/gocron/v2"
	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type OgsClient struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OgsConfig struct {
	ServerName             string `json:"server_name"`
	CdnHost                string `json:"cdn_host"`
	Cdn                    string `json:"cdn"`
	CdnRelease             string `json:"cdn_release"`
	Version                string `json:"version"`
	Release                string `json:"release"`
	BraintreeCse           string `json:"braintree_cse"`
	StripePk               string `json:"stripe_pk"`
	PaypalServer           string `json:"paypal_server"`
	PaypalEmail            string `json:"paypal_email"`
	PaypalThisServer       string `json:"paypal_this_server"`
	CsrfToken              string `json:"csrf_token"`
	SupporterCurrencyScale struct {
		ERR float64 `json:"ERR"`
		OGS float64 `json:"OGS"`
		USD float64 `json:"USD"`
		EUR float64 `json:"EUR"`
		RUB float64 `json:"RUB"`
		GBP float64 `json:"GBP"`
		CAD float64 `json:"CAD"`
		JPY float64 `json:"JPY"`
		KRW float64 `json:"KRW"`
		HKD float64 `json:"HKD"`
		CNY float64 `json:"CNY"`
		ARS float64 `json:"ARS"`
		AUD float64 `json:"AUD"`
		BRL float64 `json:"BRL"`
		BGN float64 `json:"BGN"`
		CZK float64 `json:"CZK"`
		DKK float64 `json:"DKK"`
		HUF float64 `json:"HUF"`
		ISK float64 `json:"ISK"`
		INR float64 `json:"INR"`
		MXN float64 `json:"MXN"`
		NZD float64 `json:"NZD"`
		NOK float64 `json:"NOK"`
		PLN float64 `json:"PLN"`
		RON float64 `json:"RON"`
		SGD float64 `json:"SGD"`
		SEK float64 `json:"SEK"`
		CHF float64 `json:"CHF"`
		THB float64 `json:"THB"`
		AED float64 `json:"AED"`
	} `json:"supporter_currency_scale"`
	GgsHost           string `json:"ggs_host"`
	AgaRatingsEnabled bool   `json:"aga_ratings_enabled"`
	Ogs               struct {
		Preferences struct {
			ShowGameListView bool `json:"show_game_list_view"`
		} `json:"preferences"`
	} `json:"ogs"`
	BillingMorLocations []string `json:"billing_mor_locations"`
	User                struct {
		Anonymous        bool   `json:"anonymous"`
		ID               int    `json:"id"`
		Username         string `json:"username"`
		RegistrationDate string `json:"registration_date"`
		Ratings          struct {
			Version int `json:"version"`
			Overall struct {
				Rating     float64 `json:"rating"`
				Deviation  float64 `json:"deviation"`
				Volatility float64 `json:"volatility"`
			} `json:"overall"`
		} `json:"ratings"`
		Country                string  `json:"country"`
		Professional           bool    `json:"professional"`
		Ranking                float64 `json:"ranking"`
		Provisional            float64 `json:"provisional"`
		CanCreateTournaments   bool    `json:"can_create_tournaments"`
		IsModerator            bool    `json:"is_moderator"`
		IsSuperuser            bool    `json:"is_superuser"`
		IsTournamentModerator  bool    `json:"is_tournament_moderator"`
		ModeratorPowers        float64 `json:"moderator_powers"`
		OfferedModeratorPowers float64 `json:"offered_moderator_powers"`
		Supporter              bool    `json:"supporter"`
		SupporterLevel         float64 `json:"supporter_level"`
		TournamentAdmin        bool    `json:"tournament_admin"`
		UIClass                string  `json:"ui_class"`
		Icon                   string  `json:"icon"`
		Email                  string  `json:"email"`
		EmailValidated         bool    `json:"email_validated"`
		IsAnnouncer            bool    `json:"is_announcer"`
		HasActiveWarningFlag   bool    `json:"has_active_warning_flag"`
		NeedRank               bool    `json:"need_rank"`
		StartingRankHint       string  `json:"starting_rank_hint"`
		LastSupporterTrial     string  `json:"last_supporter_trial"`
	} `json:"user"`
	ProfanityFilter bool `json:"profanity_filter"`
	Ignores         struct {
	} `json:"ignores"`
	DismissableMessages struct {
	} `json:"dismissable_messages"`
	ChatAuth         string `json:"chat_auth"`
	IncidentAuth     string `json:"incident_auth"`
	NotificationAuth string `json:"notification_auth"`
	UserJwt          string `json:"user_jwt"`
}

type EmitAuth struct {
	Auth     string `json:"auth"`
	PlayerID int    `json:"player_id"`
	Username string `json:"username"`
}

type PrivateMessage struct {
	From struct {
		ID       int     `json:"id"`
		Username string  `json:"username"`
		Ranking  float64 `json:"ranking"`
		Ratings  struct {
			Version int `json:"version"`
			Overall struct {
				Rating     float64 `json:"rating"`
				Deviation  float64 `json:"deviation"`
				Volatility float64 `json:"volatility"`
			} `json:"overall"`
		} `json:"ratings"`
		Country      string `json:"country"`
		Professional bool   `json:"professional"`
		UIClass      string `json:"ui_class"`
	} `json:"from"`
	To struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"to"`
	Message struct {
		I string `json:"i"`
		T int    `json:"t"`
		M string `json:"m"`
	} `json:"message"`
}

type PlayerDetail struct {
	Related struct {
		Challenge   string `json:"challenge"`
		Games       string `json:"games"`
		Ladders     string `json:"ladders"`
		Tournaments string `json:"tournaments"`
		Groups      string `json:"groups"`
		Icon        string `json:"icon"`
	} `json:"related"`
	ID                 int         `json:"id"`
	Username           string      `json:"username"`
	Professional       bool        `json:"professional"`
	Ranking            float64     `json:"ranking"`
	Country            string      `json:"country"`
	Language           string      `json:"language"`
	About              string      `json:"about"`
	Supporter          bool        `json:"supporter"`
	IsBot              bool        `json:"is_bot"`
	BotAi              interface{} `json:"bot_ai"`
	BotOwner           interface{} `json:"bot_owner"`
	Website            string      `json:"website"`
	RegistrationDate   time.Time   `json:"registration_date"`
	Name               interface{} `json:"name"`
	TimeoutProvisional bool        `json:"timeout_provisional"`
	Ratings            struct {
		Version int `json:"version"`
		Overall struct {
			Rating     float64 `json:"rating"`
			Deviation  float64 `json:"deviation"`
			Volatility float64 `json:"volatility"`
		} `json:"overall"`
	} `json:"ratings"`
	IsFriend bool        `json:"is_friend"`
	AgaID    interface{} `json:"aga_id"`
	UIClass  string      `json:"ui_class"`
	Icon     string      `json:"icon"`
}

type OGSchan struct {
	OgsUsername string
	UserCode    string
	ChRes       chan bool
	DiscordID   string
	GuildID     string
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func SetDiscordRole(s *discordgo.Session, name string, guildID string, rankName string) bool {
	member, err := s.GuildMember(guildID, name)
	if err != nil {
		log.Println(err)
		return false
	}
	rolesToRemove := make([]string, 0)
	x, _ := s.GuildRoles(guildID)
	var discordRole *discordgo.Role
	for _, b := range x {
		if b.Name == rankName {
			discordRole = b
		}
		pattern := `^[0-9]{1,2}(kyu|dan)\??$`
		re := regexp.MustCompile(pattern)
		for _, r := range member.Roles {
			if r == b.ID && re.MatchString(b.Name) {
				rolesToRemove = append(rolesToRemove, r)
			}
		}
	}
	if discordRole == nil {
		var color int
		if strings.Contains(rankName, "?") {
			color = 1435457
		} else if strings.Contains(rankName, "dan") {
			color = 16759040
		} else if strings.Contains(rankName, "kyu") && len(rankName) == 4 {
			color = 16464150
		} else if strings.Contains(rankName, "kyu") && len(rankName) == 5 {
			color = 35839
		} else {
			color = 1435457
		}
		role := discordgo.RoleParams{Name: rankName, Color: &color}
		var err error
		discordRole, err = s.GuildRoleCreate(guildID, &role)
		if err != nil {
			log.Println("error role create ", err)
			return false
		}
	}
	for _, r := range rolesToRemove {
		err := s.GuildMemberRoleRemove(guildID, name, r)
		if err != nil {
			log.Println(err)
		}
	}
	err = s.GuildMemberRoleAdd(guildID, name, discordRole.ID)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func UpdatePlayerRanking(name string, guildID string, conn *gosqlite.Conn, s *discordgo.Session) bool {
	ogsID := ""
	oldRank := ""
	stmt, err := conn.Prepare(`SELECT ogs_id, ranking_name FROM ogs WHERE name = ? AND guild_id = ? `, name, guildID)
	if err != nil {
		log.Println("database error select", err)
		return false
	}
	hasRow, err := stmt.Step()
	if err != nil {
		log.Println("database error step", err)
		return false
	}
	if hasRow {
		err := stmt.Scan(&ogsID, &oldRank)
		if err != nil {
			log.Fatalln(err)
			return false
		}
	}
	defer stmt.Close()
	pd := PlayerDetail{}
	err = getJson("https://online-go.com/api/v1/players/"+ogsID, &pd)
	if err != nil {
		log.Println(err)
		return false
	}
	rankName := GetRankName(pd.Ranking, pd.Ratings.Overall.Deviation)
	err = conn.Exec(`UPDATE ogs SET ogs_ranking = ?, ogs_deviation = ?, ranking_name = ? WHERE name = ? AND guild_id = ?`,
		pd.Ranking, pd.Ratings.Overall.Deviation, rankName, name, guildID)
	if err != nil {
		log.Println("database error ogs", err)
		return false
	}
	return SetDiscordRole(s, name, guildID, rankName)
}

func GetRankName(v float64, d float64) string {
	res := ""

	if v < 6.0 {
		res = "25kyu"
	} else if v < 30 {
		kyu := 30 - int(v)
		res = strconv.Itoa(kyu) + "kyu"
	} else {
		dan := (int(v) - 30) + 1
		res = strconv.Itoa(dan) + "dan"
	}

	if d >= 200.0 {
		res += "?"
	}
	return res
}

func ReadOGSPrivateMessage(s *discordgo.Session, cfg *config.Config, conn *gosqlite.Conn, scheduler *gocron.Scheduler, ogsChan OGSchan) {
	timeLimit := 120 * time.Second
	wsUrl := "wss://online-go.com/socket.io/?EIO=3&transport=websocket"
	client := &http.Client{Timeout: 10 * time.Second}
	ogsClient := OgsClient{Username: cfg.OgsUsername, Password: cfg.OgsPassword}
	authData, _ := json.MarshalIndent(&ogsClient, "", "\t")

	req, _ := http.NewRequest("POST", "https://online-go.com/api/v0/login", bytes.NewBuffer(authData))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("OGS auth error", err)
		ogsChan.ChRes <- false
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("OGS auth status code: ", resp.Status)
		return
	}
	logoutFunc := func() {
		_, err := http.Get("https://online-go.com/api/v0/logout")
		if err != nil {
			log.Println(err)
		}
	}
	defer logoutFunc()
	resp_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	var ogsConfig OgsConfig
	err = json.Unmarshal(resp_body, &ogsConfig)
	if err != nil {
		log.Println(err)
	}
	c, err := gosocketio.Dial(wsUrl, transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Println(err)
		ogsChan.ChRes <- false
		return
	}
	defer c.Close()
	err = c.Emit("authenticate", &EmitAuth{
		Auth:     ogsConfig.ChatAuth,
		Username: ogsConfig.User.Username,
		PlayerID: ogsConfig.User.ID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	ogsRanking := 0.0
	ogsDeviation := 0.0
	ogsID := 0
	ch := make(chan bool)
	defer close(ch)
	ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()
	timerDelay := 20 * time.Second
	timer := time.NewTimer(timerDelay)
	defer timer.Stop()

	pmFunc := func(i interface{}, response PrivateMessage) {
		if strings.ToLower(response.From.Username) == ogsChan.OgsUsername && response.Message.M == ogsChan.UserCode {
			ogsRanking = response.From.Ranking
			ogsDeviation = response.From.Ratings.Overall.Deviation
			ogsID = response.From.ID
			ch <- true
		}
	}
	err = c.On("private-message", pmFunc)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		select {
		case <-ch:
			t := time.Now().UTC()
			rankName := GetRankName(ogsRanking, ogsDeviation)
			err := conn.Exec(`INSERT OR REPLACE INTO ogs VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				ogsChan.DiscordID, t.Format(time.RFC1123), ogsID, ogsChan.OgsUsername, ogsRanking, ogsDeviation, rankName, ogsChan.GuildID)
			if err != nil {
				log.Fatalln("database error ogs", err)
			}
			(*scheduler).RemoveByTags(ogsChan.DiscordID + ogsChan.GuildID + "OGS")
			j, err := (*scheduler).NewJob(
				gocron.DailyJob(
					1,
					gocron.NewAtTimes(gocron.NewAtTime(uint(t.UTC().Hour()), uint(t.UTC().Minute()), uint(t.UTC().Second()))),
				),
				gocron.NewTask(UpdatePlayerRanking, ogsChan.DiscordID, ogsChan.GuildID, conn, s),
				gocron.WithTags(ogsChan.DiscordID+ogsChan.GuildID+"OGS"),
			)
			if err != nil {
				log.Fatalln(err)
			}
			_, err = j.NextRun()
			if err != nil {
				log.Fatalln(err)
			}
			if SetDiscordRole(s, ogsChan.DiscordID, ogsChan.GuildID, rankName) {
				ogsChan.ChRes <- true
			} else {
				ogsChan.ChRes <- false
			}
			return
		case <-ctx.Done():
			ogsChan.ChRes <- false
			return
		case <-timer.C:
			ping := struct {
				client int64
			}{
				client: time.Now().UnixMilli(),
			}
			err := c.Emit("net/ping", &ping)
			if err != nil {
				log.Println(err)
			}
			timer.Reset(timerDelay)
		}
	}
}

func AuthOGS(s *discordgo.Session, cfg *config.Config, conn *gosqlite.Conn, scheduler *gocron.Scheduler, ch chan OGSchan) {
	for ogsChan := range ch {
		go ReadOGSPrivateMessage(s, cfg, conn, scheduler, ogsChan)
	}
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
