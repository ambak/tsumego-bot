package command

import (
	"log"
	"math/rand"
	"os"
	"os/exec"

	"github.com/ambak/tsumego-bot/config"
	"github.com/ambak/tsumego-bot/sgf"
	"github.com/bwmarrin/discordgo"
	"github.com/eatonphil/gosqlite"
)

func Tsumego(s *discordgo.Session, m *discordgo.MessageCreate, argv []string, levels []string,
	problems [][]os.DirEntry, cfg *config.Config, conn *gosqlite.Conn, themes []string, name string, lvl string) {

	msg := ""
	level := 0
	channel := ""
	if name == "" {
		name = m.Author.ID
		channel = m.ChannelID
		level = rand.Intn(len(levels))
		lvl = levels[level]
		if len(argv) > 1 && IsValidArg(argv[1]) {
			ok := false
			for i, l := range levels {
				if argv[1] == l || argv[1] == l[:1] {
					ok = true
					lvl = l
					level = i
				}
			}
			if !ok {
				msg := "You must pass valid tsumego level. Example:\n`;tsumego advanced`"
				_, err := s.ChannelMessageSendReply(channel, msg, m.Reference())
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		}
	} else {
		msg = "Your daily tsumego. Good luck!\nTo unsubscribe type `;unsubscribe`\n"
		c, err := s.UserChannelCreate(name)
		if err != nil {
			log.Fatalln(err)
			return
		}
		channel = c.ID
		if lvl == "random" {
			level = rand.Intn(len(levels))
			lvl = levels[level]
		} else {
			for i, l := range levels {
				if lvl == l || lvl == l[:1] {
					level = i
				}
			}
		}
	}
	tsumegoName := problems[level][rand.Intn(len(problems[level]))].Name()

	theme := themes[rand.Intn(len(themes))]
	stmt, err := conn.Prepare(`SELECT theme FROM theme WHERE name = ?`, name)
	if err != nil {
		log.Println("database error select", err)
	}
	hasRow, err := stmt.Step()
	if err != nil {
		log.Println("database error step", err)
	}
	if hasRow {
		err := stmt.Scan(&theme)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer stmt.Close()
	path := cfg.Tsumego + "/" + lvl + "/" + tsumegoName
	part_rect, err := sgf.SgfSize(path)
	if err != nil {
		log.Fatalln(err)
		return
	}

	pictureName := RandStringBytesMaskImpr(10)
	tsumegoID := lvl[:1] + tsumegoName[:len(tsumegoName)-4]
	msg += "tsumegoID: `" + tsumegoID + "`"
	out := exec.Command("python3", "sgf2image/sgf2img.py", "--start", "0",
		"--end", "0", "--part_rect", part_rect, "--theme", theme, path, pictureName+".jpg")
	_, err = out.Output()
	if err != nil {
		log.Fatalln(err)
	}
	fileBytes, _ := os.Open("sgf2image/" + pictureName + ".jpg")
	_, err = s.ChannelFileSendWithMessage(channel, msg, "tsumego.jpg", fileBytes)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Remove("sgf2image/" + pictureName + ".jpg")
	if err != nil {
		log.Fatalln(err)
	}
	Solve(s, m, levels, problems, cfg, theme, tsumegoID, channel)
}
