package command

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ambak/tsumego-bot/config"
	"github.com/ambak/tsumego-bot/sgf"
	"github.com/bwmarrin/discordgo"
)

func Solve(s *discordgo.Session, m *discordgo.MessageCreate, levels []string,
	problems [][]os.DirEntry, cfg *config.Config, theme string, tsumegoID string, channel string) {
	var part_rect string
	var err error
	lvl := ""
	for _, l := range levels {
		if tsumegoID[0] == l[0] {
			lvl = l
		}
	}
	if lvl == "" {
		err = errors.New("wrong args lvl")
		log.Println(err)
	}
	part_rect, err = sgf.SgfSize(cfg.Tsumego + "/" + lvl + "/" + tsumegoID[1:] + ".sgf")
	if err != nil {
		_, err := s.ChannelMessageSend(channel, "You must pass valid tsumegoID. Example:\n`;solve a0001`")
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	pictureName := RandStringBytesMaskImpr(10)
	msg := "tsumegoID: `" + tsumegoID + "`"
	path := cfg.Tsumego + "/" + lvl + "/" + tsumegoID[1:] + ".sgf"
	out := exec.Command("python3", "sgf2image/sgf2img.py", "--start", "1", "--end", "--part_rect",
		part_rect, "--theme", theme, path, pictureName+".jpg")
	_, err = out.Output()
	if err != nil {
		log.Fatalln(err)
	}
	o := strings.Split(fmt.Sprint(out.Stdout), "\n")

	pattern := "[0-9]* = [0-9]*"
	re := regexp.MustCompile(pattern)
	for _, move := range o {
		ok := re.MatchString(move)
		if ok {
			msg += "\n||`" + move + "`||"
		}
	}

	fileBytes, _ := os.Open("sgf2image/" + pictureName + ".jpg")
	_, err = s.ChannelFileSendWithMessage(channel, msg, "SPOILER_tsumego.jpg", fileBytes)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Remove("sgf2image/" + pictureName + ".jpg")
	if err != nil {
		log.Fatalln(err)
	}
}
