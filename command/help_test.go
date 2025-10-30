package command

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHelp(t *testing.T) {
	token := os.Getenv("DISCORD_TEST_TOKEN")
	if token == "" {
		t.Fatal("DISCORD_TEST_TOKEN not set")
	}
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		t.Fatal("Cannot create Discord session:", err)
	}

	err = s.Open()
	if err != nil {
		t.Fatal("Cannot open connection:", err)
	}
	defer s.Close()
	channelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	if channelID == "" {
		t.Fatal("DISCORD_TEST_CHANNEL_ID not set")
	}
	msg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:   "help",
			ChannelID: channelID,
			Author:    &discordgo.User{ID: "1", Username: "Tester"},
		},
	}
	Help(s, msg)

	msgs, err := s.ChannelMessages(channelID, 1, "", "", "")
	if err != nil {
		t.Fatal("Get messages error:", err)
	}
	if msgs[0].ChannelID != channelID {
		t.Error("Wrong channelID:", msgs[0].ChannelID)
	}
	if msgs[0].Content != help {
		t.Error("Wrong content:", msgs[0].Content, help)
	}
}
