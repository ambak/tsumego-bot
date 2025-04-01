package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ambak/tsumego-bot/command"
	"github.com/ambak/tsumego-bot/config"
	"github.com/ambak/tsumego-bot/ogs"
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
	"github.com/go-co-op/gocron/v2"
)

var (
	cfg       config.Config
	conn      *gosqlite.Conn
	problems  [][]os.DirEntry
	themes    []string
	levels    []string
	scheduler gocron.Scheduler
	ch        chan ogs.OGSchan
)

func init() {
	configPath := flag.String("c", "config.json", "path to confing file")
	flag.Parse()
	configFile, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalln("Config not found.", err)
	}
	if err = json.Unmarshal(configFile, &cfg); err != nil {
		log.Fatalln("Config error.", err)
	}
	ch = make(chan ogs.OGSchan)
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
	err = conn.Exec(`CREATE TABLE IF NOT EXISTS theme(name TEXT, theme TEXT, PRIMARY KEY (name))`)
	if err != nil {
		log.Fatalln("database create table error", err)
		return
	}

	err = conn.Exec(`CREATE TABLE IF NOT EXISTS subscribe(name TEXT, time DATETIME, level TEXT, PRIMARY KEY (name))`)
	if err != nil {
		log.Fatalln("database create table error", err)
		return
	}

	err = conn.Exec(`CREATE TABLE IF NOT EXISTS ogs(name TEXT, time DATETIME, ogs_id TEXT, ogs_username TEXT, 
					ogs_ranking REAL, ogs_deviation REAL, ranking_name TEXT, guild_id TEXT, PRIMARY KEY (name, guild_id))`)
	if err != nil {
		log.Fatalln("database create table error", err)
		return
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatalln("error creating Discord session,", err)
		return
	}
	defer dg.Close()
	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages
	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection,", err)
		return
	}

	scheduler, _ = gocron.NewScheduler(gocron.WithLocation(time.UTC))
	defer func() { _ = scheduler.Shutdown() }()
	stmt, err := conn.Prepare(`SELECT * FROM subscribe`)
	if err != nil {
		log.Println("database error subscribe init", err)
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Println("database error subscribe", err)
			return
		}
		if !hasRow {
			break
		}
		var name string
		var ttime string
		var level string
		err = stmt.Scan(&name, &ttime, &level)
		if err != nil {
			log.Fatalln(err)
		}
		t, _ := time.Parse(time.RFC1123, ttime)
		_, err = scheduler.NewJob(
			gocron.DailyJob(
				1,
				gocron.NewAtTimes(gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second()))),
			),
			gocron.NewTask(command.Tsumego, dg, &discordgo.MessageCreate{}, []string{}, levels, problems, &cfg, conn, themes, name, level),
			gocron.WithTags(name),
		)
		if err != nil {
			log.Fatalln(err)
		}
	}
	stmt, err = conn.Prepare(`SELECT name, time, guild_id FROM ogs`)
	if err != nil {
		log.Println("database error subscribe init", err)
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Println("database error ogs", err)
			return
		}
		if !hasRow {
			break
		}
		var name string
		var ttime string
		var guildID string
		err = stmt.Scan(&name, &ttime, &guildID)
		if err != nil {
			log.Fatalln(err)
		}
		t, _ := time.Parse(time.RFC1123, ttime)
		_, err = scheduler.NewJob(
			gocron.DailyJob(
				1,
				gocron.NewAtTimes(gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second()))),
			),
			gocron.NewTask(ogs.UpdatePlayerRanking, name, guildID, conn, dg),
			gocron.WithTags(name+guildID+"OGS"),
		)
		if err != nil {
			log.Fatalln(err)
		}
	}
	scheduler.Start()

	go ogs.AuthOGS(dg, &cfg, conn, &scheduler, ch)

	log.Println("Bot is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Fatalln("App EXIT")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content[0] == ';' {
		argv := strings.Fields(m.Content)
		if argv[0] == ";tsumego" || argv[0] == ";t" {
			command.Tsumego(s, m, argv, levels, problems, &cfg, conn, themes, "", "")
		} else if argv[0] == ";theme" {
			command.Theme(s, m, argv, themes, conn)
		} else if argv[0] == ";level" || argv[0] == ";l" {
			command.Level(s, m, levels)
		} else if argv[0] == ";randomtheme" {
			command.Randomtheme(s, m, conn)
		} else if argv[0] == ";subscribe" {
			command.Subscribe(s, m, conn, argv, levels, problems, &cfg, themes, &scheduler)
		} else if argv[0] == ";unsubscribe" {
			command.Unsubscribe(s, m, conn, &scheduler)
		} else if argv[0] == ";link" {
			command.Link(s, m, conn, argv, &cfg, &scheduler, ch)
		} else {
			command.Help(s, m)
		}
	}
}
