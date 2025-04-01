package command

import (
	"log"
	"strings"

	"github.com/ambak/tsumego-bot/config"
	"github.com/ambak/tsumego-bot/ogs"
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
	"github.com/go-co-op/gocron/v2"
)

func Link(s *discordgo.Session, m *discordgo.MessageCreate, conn *gosqlite.Conn, argv []string,
	cfg *config.Config, scheduler *gocron.Scheduler, ch chan ogs.OGSchan) {

	if m.GuildID == "" {
		msg := "You cannot link your OGS account in private message."
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	if len(argv) < 2 {
		msg := "You must pass your OGS username. Example:\n`;link YOUR_USERNAME`"
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	ogsUsername := m.Content[6:]
	userCode := RandStringBytesMaskImpr(5)
	msg := "Send a private message to `" + cfg.OgsUsername + "` on OGS within 120 seconds.\n"
	msg += cfg.OgsUrl + "\n"
	msg += "Message content:\n`" + userCode + "`"
	_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
	if err != nil {
		log.Fatalln(err)
	}
	chRes := make(chan bool)
	defer close(chRes)

	ogsChan := ogs.OGSchan{
		OgsUsername: strings.ToLower(ogsUsername),
		UserCode:    userCode,
		ChRes:       chRes,
		DiscordID:   m.Author.ID,
		GuildID:     m.GuildID,
	}
	ch <- ogsChan

	res := <-chRes
	if res {
		msg := "You have successfully linked your OGS account."
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		msg := "Something went wrong, please try again."
		_, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		if err != nil {
			log.Fatalln(err)
		}
	}
}
