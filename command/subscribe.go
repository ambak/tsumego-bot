package command

import (
	"log"
	"os"
	"time"

	"github.com/ambak/tsumego-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
	"github.com/go-co-op/gocron/v2"
)

func Subscribe(s *discordgo.Session, m *discordgo.MessageCreate, conn *gosqlite.Conn, argv []string,
	levels []string, problems [][]os.DirEntry, cfg *config.Config, themes []string, scheduler *gocron.Scheduler) {
	level := "random"
	if len(argv) > 1 && IsValidArg(argv[1]) {
		ok := false
		for _, l := range levels {
			if argv[1] == l || argv[1] == l[:1] {
				ok = true
				level = l
			}
		}
		if !ok {
			msg := "You must pass valid tsumego level. Example:\n`;subscribe advanced`"
			_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
			if err != nil {
				log.Fatalln(err)
			}
			return
		}
	}
	t := time.Now().UTC()
	err := conn.Exec(`INSERT OR REPLACE INTO subscribe VALUES (?, ?, ?)`, m.Author.ID, t.Format(time.RFC1123), level)
	if err != nil {
		log.Fatalln("database error subscribe", err)
	}
	(*scheduler).RemoveByTags(m.Author.ID)
	j, err := (*scheduler).NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(gocron.NewAtTime(uint(t.UTC().Hour()), uint(t.UTC().Minute()), uint(t.UTC().Second()))),
		),
		gocron.NewTask(Tsumego, s, &discordgo.MessageCreate{}, []string{}, levels, problems, cfg, conn, themes, m.Author.ID, level),
		gocron.WithTags(m.Author.ID),
	)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = j.NextRun()
	if err != nil {
		log.Fatalln(err)
	}

	msg := "Thanks for subscribing.\n" +
		"Every day you will receive a direct message with a new " +
		"`" + level + "`" + " level tsumego " +
		"\nTo unsubscribe type `;unsubscribe`"
	_, err = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
	if err != nil {
		log.Fatalln(err)
	}
	Tsumego(s, &discordgo.MessageCreate{}, []string{}, levels, problems, cfg, conn, themes, m.Author.ID, level)
}

func Unsubscribe(s *discordgo.Session, m *discordgo.MessageCreate, conn *gosqlite.Conn, scheduler *gocron.Scheduler) {
	err := conn.Exec(`DELETE FROM subscribe WHERE name = ?`, m.Author.ID)
	if err != nil {
		log.Println("database error subscribe", err)
	}
	(*scheduler).RemoveByTags(m.Author.ID)
	msg := "Bye-bye"
	_, err = s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
	if err != nil {
		log.Fatalln(err)
	}
}
