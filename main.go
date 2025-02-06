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
	err = conn.Exec(`CREATE TABLE IF NOT EXISTS subscribe(name TEXT, time DATETIME, level TEXT, PRIMARY KEY (name))`)
	if err != nil {
		log.Fatalln("database create table error", err)
		return
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	defer dg.Close()
	if err != nil {
		log.Fatalln("error creating Discord session,", err)
		return
	}
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
		t, _ := time.Parse(time.RFC1123, ttime)
		scheduler.NewJob(
			gocron.DailyJob(
				1,
				gocron.NewAtTimes(gocron.NewAtTime(uint(t.UTC().Hour()), uint(t.UTC().Minute()), uint(t.UTC().Second()))),
			),
			gocron.NewTask(command.Tsumego, dg, &discordgo.MessageCreate{}, []string{}, levels, problems, &cfg, conn, themes, name, level),
			gocron.WithTags(name),
		)
	}
	scheduler.Start()

	log.Println("Bot is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
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
		} else {
			command.Help(s, m)
		}
	}
}
