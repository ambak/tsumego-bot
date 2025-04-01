package command

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

const help = "```" + `
Available commands:
	;tsumego                Show random tsumego. Shortcut ;t

	;tsumego LVL            Show tsumego at level LVL

	;level                  Show available levels. Shortcut LVL by using first character

	;theme                  Show available themes

	;theme name             Select your theme

	;randomtheme            Set theme to random

	;subscribe              Subscribe to get daily tsumego

	;subscribe LVL          Subscribe to get daily tsumego at level LVL

	;unsubscribe            Unsubscribe daily tsumego

	;link OGS_USERNAME      Link your OSG account and your discord account

Example:
	;t e                    Show elementary tsumego

	;theme cartoon          Set theme to "cartoon"

	;subscribe a            Subscribe to daily tsumego at advanced level
	` + "```"

func Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, help)
	if err != nil {
		log.Fatalln(err)
	}
}
