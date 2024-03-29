package logic

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/bwmarrin/discordgo"
)

func detectBusts(s *discordgo.Session, m *discordgo.MessageCreate) {
	if isSiup(m.Message) {
		onSiupServer(s, m)
		return
	}

	if isBump(m.Message) {
		onBumpServer(s, m)
		return
	}

	if isLike(m.Message) {
		onLikeServer(s, m)
		return
	}
}

func onSiupServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.LastSiup = time.Now()
	go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind), "S.up")
	fmt.Println("FOUND S.up")

	for _, user := range chMonitorWriters {
		if user.strify == m.Interaction.Member.User.ID {
			fmt.Println("FOUND S.up user", user.id)

			sendAndLog(s, user.id, "S.up", config.SumForPaying)
			return
		}
	}
}

func onBumpServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.LastBump = time.Now()
	go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind), "Bump")
	fmt.Println("FOUND Bump")

	userID := strings.Split(strings.Split(m.Embeds[0].Description, "<@")[1], ">")[0]
	if len(userID) == 0 {
		fmt.Println("ERROR Bump get user ID:", m.Embeds[0].Description)
		return
	}

	if strings.HasPrefix(userID, "!") {
		userID = userID[1:]
	}

	fmt.Println("FOUND Bump user", userID)

	sendAndLog(s, userID, "Bump", config.SumForPaying)
}

func onLikeServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.LastLike = time.Now()

	go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind), "Like")
	fmt.Println("FOUND Like")

	if m.Interaction != nil {
		fmt.Println("FOUND Like user", m.Interaction.User.ID)

		sendAndLog(s, m.Interaction.User.ID, "Like", config.SumForPaying)
		return
	}

	for _, user := range chMonitorWriters {
		if user.strify == m.Embeds[0].Author.Name {
			fmt.Println("FOUND Like user", user.id)

			sendAndLog(s, user.id, "Like", config.SumForPaying)
			return
		}
	}
}

func sendAndLog(s *discordgo.Session, userID string, str string, sum string) {
	if rand.Intn(100) == 0 {
		sum = sum + "0"
	}

	isum, err := strconv.Atoi(sum)
	if err != nil {
		fmt.Println("ERROR "+str+" parsing sum:", err)
	}
	_, err = config.API.AddToBalance(config.GdHouseID, userID, 0, isum, "for "+str)
	if err != nil {
		fmt.Println("ERROR "+str+" updating user balance:", err)

		_, err = s.ChannelMessageSendComplex(config.ChForLogsID, &discordgo.MessageSend{
			Content: "<@" + config.UsMainWorker + ">",
			Embed: &discordgo.MessageEmbed{
				Color:       16711680,
				Description: "Кажись, что-то пошло не так... <@" + userID + "> сделал " + str + ", но денег ему не дали(",
			},
		})
		if err != nil {
			fmt.Println("ERROR "+str+" sending wrong report message:", err)
		}
		_, err = s.ChannelMessageSendEmbed(config.ChForBustsID, &discordgo.MessageEmbed{
			Color:       16711680,
			Description: "<@" + userID + ">, у нас снова что-то сломалось, но не волнуйтесь - деньги Вам прилетят чуть позже)" + " Обратитесь за конпенсацией к Главному Разработчику",
		})
		if err != nil {
			fmt.Println("ERROR "+str+" sending wrong log message:", err)
		}

		return
	}

	_, err = s.ChannelMessageSendEmbed(config.ChForLogsID, &discordgo.MessageEmbed{
		Color:       255255,
		Description: sum + "<:AH_AniCoin:579712087224483850> были выданы <@" + userID + ">, за то что он сделал " + str,
	})
	if err != nil {
		fmt.Println("ERROR "+str+" sending right report message:", err)
	}
	_, err = s.ChannelMessageSendEmbed(config.ChForBustsID, &discordgo.MessageEmbed{
		Color:       255255,
		Description: "<@" + userID + ">, " + fmt.Sprintf(config.Responces[rand.Intn(len(config.Responces))], str, sum+"<:AH_AniCoin:579712087224483850>"),
	})
	if err != nil {
		fmt.Println("ERROR "+str+" sending right log message:", err)
	}

	fmt.Println("GUILD "+str+" by  ", userID)
}
