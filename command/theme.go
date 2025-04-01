package command

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
)

func Theme(s *discordgo.Session, m *discordgo.MessageCreate, argv []string, themes []string, conn *gosqlite.Conn) {
	if argv[0] == ";theme" {
		newTheme := ""
		if len(argv) > 1 {
			for _, t := range themes {
				if argv[1] == t {
					newTheme = t
				}
			}
			if newTheme != "" {
				err := conn.Exec(`INSERT OR REPLACE INTO theme VALUES (?, ?)`, m.Author.ID, newTheme)
				if err != nil {
					log.Fatalln(err)
				}
				s.State.User.Mention()
				_, err = s.ChannelMessageSendReply(m.ChannelID, "Your default theme is set to `"+newTheme+"`", m.Reference())
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
		if newTheme == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "```\n"+strings.Join(themes, "\n")+"```")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
