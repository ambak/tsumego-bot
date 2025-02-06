package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
)

func Randomtheme(s *discordgo.Session, m *discordgo.MessageCreate, conn *gosqlite.Conn) {
	conn.Exec(`DELETE FROM theme WHERE name = ?`, m.Author.ID)
	s.ChannelMessageSendReply(m.ChannelID, "Your theme is set to `random`", m.Reference())
}
