package logic

import (
	"fmt"
	"strings"
	"time"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/SteMak/ani-helper/workerTools/database"
	"github.com/bwmarrin/discordgo"
)

var (
	chMonitorWriters []simplifiedUser
)

type simplifiedUser struct {
	id     string
	strify string
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID == config.GdHouseID {
		if m.Content == "R U TYT?" && m.ChannelID == config.ChForLogsID {
			s.ChannelMessageSend(m.ChannelID, "I TYT KUSHAU!")
			return
		}

		if m.ChannelID == config.ChForBustsID && len(m.Content) >= 10 && strings.ToLower(m.Content[:10]) == "когда" {
			sendHelpBustMessage(s, m)
			return
		}

		if m.ChannelID == config.ChForBustsID {
			if len(chMonitorWriters) >= 30 {
				chMonitorWriters = chMonitorWriters[1:]
			}

			chMonitorWriters = append(chMonitorWriters, simplifiedUser{
				id:     m.Author.ID,
				strify: m.Author.String(),
			})
		}

		if m.ChannelID == config.ChForBustsID && len(m.Embeds) > 0 {
			detectBusts(s, m)
			return
		}

		isRequest, request := checkRequest(m.Content)
		if isRequest {
			member, err := s.GuildMember(config.GdHouseID, m.Author.ID)
			if err != nil {
				return
			}
			if len(member.Roles) > 0 && hasRole(member, config.RoRequestMakerID) {
				detectRequest(s, m.ChannelID, "-запрос"+request, m.Author)
				return
			}
		}
	}
}

func reactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.ChannelID == config.ChForRequestID && r.UserID == config.UsConfirmatorID {
		item, err := database.Records.Record(r.MessageID)
		if err != nil {
			return
		}

		fmt.Println("FOUND "+item.EmbedID+" reation added", r.Emoji.Name)

		emojiOnRequest(s, r, item)

		fmt.Println("GUILD " + item.EmbedID + " request processed successfuly")
	}
}

func onStart(s *discordgo.Session) {
	defineLastBusts(s)
	checkAndRemind(s)

	sleepSiup := int64(config.TimeWaitSiup)*60 - (time.Now().Unix() - config.LastSiup.Unix()) - int64(config.TimeRemind)*60
	sleepBump := int64(config.TimeWaitBump)*60 - (time.Now().Unix() - config.LastBump.Unix()) - int64(config.TimeRemind)*60
	sleepLike := int64(config.TimeWaitLike)*60 - (time.Now().Unix() - config.LastLike.Unix()) - int64(config.TimeRemind)*60

	if sleepSiup > 0 {
		go sleep(s, sleepSiup)
	}
	if sleepBump > 0 {
		go sleep(s, sleepBump)
	}
	if sleepLike > 0 {
		go sleep(s, sleepLike)
	}
}

func sleep(s *discordgo.Session, sleep int64) {
	time.Sleep(time.Duration(sleep) * time.Second)
	checkAndRemind(s)
}
