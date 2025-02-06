package command

import "github.com/bwmarrin/discordgo"

const help = "```" + `
Available commands:
	;tsumego                Show random tsumego. Shortcut ;t

	;tsumego lvl            Show tsumego at level lvl

	;level                  Show available levels. Shortcut lvl by using first character

	;theme                  Show available themes

	;theme name             Select your theme

	;randomtheme            Set theme to random

	;subscribe              Subscribe to get daily tsumego

	;subscribe lvl          Subscribe to get daily tsumego at level lvl

	;unsubscribe            Unsubscribe daily tsumego

Example:
	;t e                    Show elementary tsumego

	;theme cartoon          Set theme to "cartoon"

	;subscribe a            Subscribe to daily tsumego at advanced level
	` + "```"

func Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, help)
}
