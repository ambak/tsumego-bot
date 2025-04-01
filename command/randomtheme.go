package command

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
)

func Randomtheme(s *discordgo.Session, m *discordgo.MessageCreate, conn *gosqlite.Conn) {
	err := conn.Exec(`DELETE FROM theme WHERE name = ?`, m.Author.ID)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = s.ChannelMessageSendReply(m.ChannelID, "Your theme is set to `random`", m.Reference())
	if err != nil {
		log.Fatalln(err)
	}
}
