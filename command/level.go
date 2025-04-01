package command

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Level(s *discordgo.Session, m *discordgo.MessageCreate, levels []string) {
	_, err := s.ChannelMessageSend(m.ChannelID, "```\n"+strings.Join(levels, "\n")+"```")
	if err != nil {
		log.Fatalln(err)
	}
}
