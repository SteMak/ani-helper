package logic

import (
	"fmt"
	"time"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/bwmarrin/discordgo"
)

func defineLastBusts(s *discordgo.Session) {
	beforeID := ""
	var timeZero time.Time
	for config.LastBump == timeZero || config.LastSiup == timeZero || config.LastLike == timeZero {
		mess, err := s.ChannelMessages(config.ChForBustsID, 100, beforeID, "", "")
		if err != nil {
			fmt.Println("ERROR getting messages", err)
			continue
		}

		for _, m := range mess {
			if isSiup(m) && config.LastSiup == timeZero {
				config.LastSiup, _ = m.Timestamp.Parse()
			}
			if isBump(m) && config.LastBump == timeZero {
				config.LastBump, _ = m.Timestamp.Parse()
			}
			if isLike(m) && config.LastLike == timeZero {
				config.LastLike, _ = m.Timestamp.Parse()
			}
			if isRemind(m) && config.LastRemind == timeZero {
				config.LastRemind, _ = m.Timestamp.Parse()
			}
		}
		beforeID = mess[len(mess)-1].ID
	}
}

func checkAndRemind(s *discordgo.Session, str string) {
	if str == "Like" {
		if time.Now().Unix() >= config.LastLike.Unix()+int64(config.TimeWaitLike*60-config.TimeRemind) {
			remind(s, "Like")
			return
		}
	}
	if str == "Bump" {
		if time.Now().Unix() >= config.LastBump.Unix()+int64(config.TimeWaitBump*60-config.TimeRemind) {
			remind(s, "Bump")
			return
		}
	}
	if str == "S.up" {
		if time.Now().Unix() >= config.LastSiup.Unix()+int64(config.TimeWaitSiup*60-config.TimeRemind) {
			remind(s, "S.up")
			return
		}
	}
}

func remind(s *discordgo.Session, str string) {
	if time.Now().Unix() >= config.LastRemind.Unix()+int64(config.TimeDoubleRemind)*60 {
		config.LastRemind = time.Now()
		_, err := s.ChannelMessageSend(config.ChForBustsID, "<@&"+config.RoBuster+">, Скоро будет "+str)
		if err != nil {
			fmt.Println("ERROR Remind send:", err)
			return
		}
		fmt.Println("PRINT Remind " + str)
	} else {
		_, err := s.ChannelMessageSend(config.ChForBustsID, "Скоро будет "+str)
		if err != nil {
			fmt.Println("ERROR ShadowRemind send:", err)
			return
		}
		fmt.Println("PRINT ShadowRemind " + str)
	}

	go func(s *discordgo.Session, str string) {
		if str == "Like" {
			time.Sleep(time.Duration(int64(config.TimeWaitLike)*60-(time.Now().Unix()-config.LastLike.Unix())-1) * time.Second)
		}
		if str == "Bump" {
			time.Sleep(time.Duration(int64(config.TimeWaitBump)*60-(time.Now().Unix()-config.LastBump.Unix())-1) * time.Second)
		}
		if str == "S.up" {
			time.Sleep(time.Duration(int64(config.TimeWaitSiup)*60-(time.Now().Unix()-config.LastSiup.Unix())-1) * time.Second)
		}
		_, err := s.ChannelMessageSend(config.ChForBustsID, str+" заряжен!")
		if err != nil {
			fmt.Println("ERROR LastRemind send:", err)
			return
		}
		fmt.Println("PRINT LastRemind " + str)
	}(s, str)
}

func sendHelpBustMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	text := "```cs\n"
	sleepSiup := int64(config.TimeWaitSiup)*60 - (time.Now().Unix() - config.LastSiup.Unix())
	text += secondsToString("S.up", sleepSiup)
	sleepBump := int64(config.TimeWaitBump)*60 - (time.Now().Unix() - config.LastBump.Unix())
	text += secondsToString("Bump", sleepBump)
	sleepLike := int64(config.TimeWaitLike)*60 - (time.Now().Unix() - config.LastLike.Unix())
	text += secondsToString("Like", sleepLike)
	text += "\n```"

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: m.Author.AvatarURL("128"),
			Name:    m.Author.Username + "#" + m.Author.Discriminator,
		},
		Description: text,
		Color:       5869507,
	})
	if err != nil {
		fmt.Println("ERROR sending helping embed", err)
	}
}
