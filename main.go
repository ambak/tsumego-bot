package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
)

type config struct {
	Token    string `json:"token"`
	LogFile  string `json:"log_file"`
	Database string `json:"database"`
	Tsumego  string `json:"tsumego"`
}

var (
	cfg  config
	conn *gosqlite.Conn
)

func init() {
	configPath := flag.String("c", "config.json", "path to confing file")
	flag.Parse()
	configFile, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalln("Config not found.", err)
		panic("Config not found.")
	}
	if err = json.Unmarshal(configFile, &cfg); err != nil {
		log.Fatalln("Config error.", err)
		panic("Config error.")
	}
}

var problems [][]os.DirEntry
var themes []string
var levels []string

const help = "```" + `
Available commands:
	;tsumego                Show random tsumego. Shortcut ;t

	;tsumego lvl            Show tsumego at level lvl

	;level                  Show available levels. Shortcut lvl by using first character

	;solve tsumegoID        Show solution to tsumegoID. Shortcut ;s

	;theme                  Show available themes

	;theme name             Select your theme

	;randomtheme            Set theme to random
	` + "```"

func sgfSize(name string) (string, error) {
	f, err := os.ReadFile(name)
	if err != nil {
		return "0,0,19,19", err
	}
	s := string(f)
	minfirst := byte('s')
	maxfirst := byte('a')
	minsecond := byte('s')
	maxsecond := byte('a')
	setblack := false
	setwhite := false
	sz := 19
	for i := 0; i < len(s)-6; i++ {
		if s[i:i+3] == "SZ[" {
			sz, err = strconv.Atoi(s[i+3 : i+5])
			if err != nil {
				sz = 19
			}
		}
		if (!setblack || !setwhite) && s[i] == 'A' && (s[i+1] == 'B' || s[i+1] == 'W') && s[i+2] == '[' {
			if s[i+1] == 'B' {
				setblack = true
			} else {
				setwhite = true
			}
			j := i + 2
			for s[j] == '[' && s[j+3] == ']' {
				if s[j+1] >= 'a' && s[j+1] <= 's' {
					minfirst = min(minfirst, s[j+1])
					maxfirst = max(maxfirst, s[j+1])
				}
				if s[j+2] >= 'a' && s[j+2] <= 's' {
					minsecond = min(minsecond, s[j+2])
					maxsecond = max(maxsecond, s[j+2])
				}
				j += 4
			}
			i = j - 1
		}
		if s[i] == ';' && (s[i+1] == 'B' || s[i+1] == 'W') {
			if s[i+2] == '[' && s[i+5] == ']' {
				if s[i+3] >= 'a' && s[i+3] <= 's' {
					minfirst = min(minfirst, s[i+3])
					maxfirst = max(maxfirst, s[i+3])
				}
				if s[i+4] >= 'a' && s[i+4] <= 's' {
					minsecond = min(minsecond, s[i+4])
					maxsecond = max(maxsecond, s[i+4])
				}
			}
		}
	}
	left, top, right, bottom := 0, 0, sz, sz
	left = max(left, int(minfirst-byte('a'))-3)
	top = max(top, int(minsecond-byte('a'))-3)
	right = min(right, int(maxfirst-byte('a'))+3)
	bottom = min(bottom, int(maxsecond-byte('a'))+3)
	return strconv.Itoa(left) + "," + strconv.Itoa(sz+1-bottom) + "," + strconv.Itoa(right) + "," + strconv.Itoa(sz-top), nil
}

func isValidArg(name string) bool {
	ok, err := regexp.MatchString("[aei][-a-z0-9A-Z_]*", name)
	if err != nil {
		return false
	}
	return ok
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(logFile)
	log.SetOutput(mw)
	dg, err := discordgo.New("Bot " + cfg.Token)
	defer dg.Close()
	if err != nil {
		log.Fatalln("error creating Discord session,", err)
		return
	}
	levels = append(levels, "elementary")
	levels = append(levels, "intermediate")
	levels = append(levels, "advanced")
	for _, l := range levels {
		problem, err := os.ReadDir(cfg.Tsumego + "/" + l)
		if err != nil {
			log.Fatalln("error loading tsumego,", err)
			return
		}
		problems = append(problems, problem)
	}

	themes_list, err := os.ReadDir("sgf2image/themes")
	if err != nil {
		log.Fatalln("error loading themes,", err)
		return
	}
	themes = []string{}
	for _, t := range themes_list {
		if t.IsDir() {
			themes = append(themes, t.Name())
		}
	}

	conn, err = gosqlite.Open(cfg.Database)
	if err != nil {
		log.Fatalln("database load error", err)
		return
	}
	defer conn.Close()
	err = conn.Exec(`CREATE TABLE IF NOT EXISTS users(name TEXT, theme TEXT, PRIMARY KEY (name))`)
	if err != nil {
		log.Fatalln("database create table error", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection,", err)
		return
	}

	log.Println("Bot is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Fatalln("App EXIT")
}

type Gopher struct {
	Name string `json:"name"`
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content[0] == ';' {
		theme := themes[rand.Intn(len(themes))]
		stmt, err := conn.Prepare(`SELECT theme FROM users WHERE name = ?`, m.Author.ID)
		if err != nil {
			log.Println("database error select", err)
		}
		hasRow, err := stmt.Step()
		if err != nil {
			log.Println("database error step", err)
		}
		if hasRow {
			stmt.Scan(&theme)
		}
		argv := strings.Fields(m.Content)
		if argv[0] == ";tsumego" || argv[0] == ";t" {
			level := rand.Intn(len(levels))
			lvl := levels[level]
			if len(argv) > 1 && isValidArg(argv[1]) {
				ok := false
				for i, l := range levels {
					if argv[1] == l || argv[1] == l[:1] {
						ok = true
						lvl = l
						level = i
					}
				}
				if !ok {
					s.ChannelMessageSend(m.ChannelID, "You must pass valid tsumego level. Example:\n`;tsumego advanced`")
					return
				}
			}
			tsumegoName := problems[level][rand.Intn(len(problems[level]))].Name()
			path := cfg.Tsumego + "/" + lvl + "/" + tsumegoName
			part_rect, err := sgfSize(path)
			if err != nil {
				log.Fatalln(err)
				return
			}
			tsumegoID := lvl[:1] + tsumegoName[:len(tsumegoName)-4]
			out := exec.Command("python3", "sgf2image/sgf2img.py", "--start", "0",
				"--end", "0", "--part_rect", part_rect, "--theme", theme, path, "tmp"+m.ID+".jpg")
			out.Output()
			fileBytes, _ := os.Open("sgf2image/tmp" + m.ID + ".jpg")
			s.ChannelFileSendWithMessage(m.ChannelID, "tsumegoID: `"+tsumegoID+"`", "tsumego.jpg", fileBytes)
			err = os.Remove("sgf2image/tmp" + m.ID + ".jpg")
			if err != nil {
				log.Fatal(err)
			}
		} else if argv[0] == ";solve" || argv[0] == ";s" {
			var part_rect string
			var err error
			lvl := ""
			if len(argv) > 1 && isValidArg(argv[1]) {
				for _, l := range levels {
					if argv[1][0] == l[0] {
						lvl = l
					}
				}
				if lvl == "" {
					err = errors.New("wrong args lvl")
				}
				part_rect, err = sgfSize(cfg.Tsumego + "/" + lvl + "/" + argv[1][1:] + ".sgf")
			} else {
				err = errors.New("wrong args")
			}
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "You must pass valid tsumegoID. Example:\n`;solve a0001`")
				return
			}
			msg := "tsumegoID: `" + argv[1] + "`"
			path := cfg.Tsumego + "/" + lvl + "/" + argv[1][1:] + ".sgf"
			out := exec.Command("python3", "sgf2image/sgf2img.py", "--start", "1", "--end", "--part_rect",
				part_rect, "--theme", theme, path, "tmp"+m.ID+".jpg")
			out.Output()
			o := strings.Split(fmt.Sprint(out.Stdout), "\n")
			for _, move := range o {
				ok, err := regexp.MatchString("[0-9]* = [0-9]*", move)
				if err != nil {
					continue
				}
				if ok {
					msg += "\n||`" + move + "`||"
				}
			}
			fileBytes, _ := os.Open("sgf2image/tmp" + m.ID + ".jpg")
			s.ChannelFileSendWithMessage(m.ChannelID, msg, "SPOILER_tsumego.jpg", fileBytes)
			err = os.Remove("sgf2image/tmp" + m.ID + ".jpg")
			if err != nil {
				log.Fatal(err)
			}
		} else if argv[0] == ";theme" {
			newTheme := ""
			if len(argv) > 1 {
				for _, t := range themes {
					if argv[1] == t {
						newTheme = t
					}
				}
				if newTheme != "" {
					conn.Exec(`INSERT OR REPLACE INTO users VALUES (?, ?)`, m.Author.ID, newTheme)
					s.State.User.Mention()
					s.ChannelMessageSendReply(m.ChannelID, "Your default theme is set to `"+newTheme+"`", m.Reference())
				}
			}
			if newTheme == "" {
				s.ChannelMessageSend(m.ChannelID, "```\n"+strings.Join(themes, "\n")+"```")
			}
		} else if argv[0] == ";level" || argv[0] == ";l" {
			s.ChannelMessageSend(m.ChannelID, "```\n"+strings.Join(levels, "\n")+"```")
		} else if argv[0] == ";randomtheme" {
			conn.Exec(`DELETE FROM users WHERE name = ?`, m.Author.ID)
			s.ChannelMessageSendReply(m.ChannelID, "Your theme is set to `random`", m.Reference())
		} else {
			s.ChannelMessageSend(m.ChannelID, help)
		}
	}
}
