package command

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Level(s *discordgo.Session, m *discordgo.MessageCreate, levels []string) {
	s.ChannelMessageSend(m.ChannelID, "```\n"+strings.Join(levels, "\n")+"```")
}
