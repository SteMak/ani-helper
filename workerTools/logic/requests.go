package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/SteMak/ani-helper/workerTools/database"
	"github.com/bwmarrin/discordgo"
)

func createEmbed(s *discordgo.Session, item *database.Record, color int) *discordgo.MessageEmbed {
	var users string

	uss := database.Uss{}
	json.Unmarshal([]byte(item.UsersSum), &uss)
	for _, userSum := range uss.Us {
		users += strconv.FormatUint(userSum.Sum, 10) + "\n<@" + strings.Join(userSum.Users, "> <@") + ">\n"

		if color == 255255 {
			for _, user := range userSum.Users {
				_, err = config.API.AddToBalance(config.GdHouseID, user, 0, int(userSum.Sum), item.Reason)
				if err != nil {
					fmt.Println("ERROR "+item.EmbedID+" add money to user balance:", err)
					_, err = s.ChannelMessageSendEmbed(config.ChForInfoRequestID, &discordgo.MessageEmbed{
						Color:       16711680,
						Description: "Кажись, что-то пошло не так... <@" + user + "> не получил денег за " + item.Reason + ". Обратитесь за конпенсацией к Главному Разработчику",
					})
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending wrong info message:", err)
					}
					_, err = s.ChannelMessageSendComplex(config.ChForLogsID, &discordgo.MessageSend{
						Content: "<@" + config.UsMainWorker + ">",
						Embed: &discordgo.MessageEmbed{
							Color:       16711680,
							Description: "Кажись, что-то пошло не так... " + strconv.FormatUint(userSum.Sum, 10) + "<:AH_AniCoin:579712087224483850> не были выданы <@" + user + ">, за " + item.Reason,
						},
					})
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending wrong report message:", err)
					}
				} else {
					_, err = s.ChannelMessageSendEmbed(config.ChForInfoRequestID, &discordgo.MessageEmbed{
						Color:       255255,
						Description: "<@" + user + ">, " + "Вы получили " + strconv.FormatUint(userSum.Sum, 10) + "<:AH_AniCoin:579712087224483850>. Причина: " + item.Reason,
					})
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending right info message:", err)
					}
					_, err = s.ChannelMessageSendEmbed(config.ChForLogsID, &discordgo.MessageEmbed{
						Color:       255255,
						Description: strconv.FormatUint(userSum.Sum, 10) + "<:AH_AniCoin:579712087224483850> были выданы <@" + user + ">, за " + item.Reason,
					})
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending right report message:", err)
					}
				}
			}
		}
	}

	return &discordgo.MessageEmbed{
		Title: "Запрос на выдачу денег",
		Author: &discordgo.MessageEmbedAuthor{
			Name:    item.AuthorName,
			IconURL: item.AuthorIcon,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Запрос",
				Value: item.Reason,
			},
			{
				Name:  "Получатели",
				Value: users,
			},
		},
		Color: color,
	}
}

func detectRequest(s *discordgo.Session, channelID, content string, author *discordgo.User) {
	if content == "-запрос" {
		s.ChannelMessageSend(channelID, config.Usage)
		return
	}

	if !strings.HasPrefix(content, "-запрос ") {
		return
	}

	item, err := queryIntoRecord(s, channelID, content, author)
	if err != nil {
		fmt.Println("BADQR parsing query:", err)
		return
	}

	err = database.Records.New(item)
	if err != nil {
		fmt.Println("ERROR create new database record", err)
		return
	}

	_, err = s.ChannelMessageSend(channelID, "Запрос отправлен")
	if err != nil {
		fmt.Println("ERROR sending confirm message", err)
	}

	fmt.Println("GUILD " + item.EmbedID + " request sended")
}

func queryIntoRecord(s *discordgo.Session, channelID, content string, author *discordgo.User) (*database.Record, error) {
	item, err := parseQuery(s, channelID, content, author)
	if err != nil {
		return item, err
	}

	fmt.Println("FOUND request message", content)

	err = resendRequest(s, channelID, content, item)
	if err != nil {
		return item, err
	}

	return item, nil
}

func parseQuery(s *discordgo.Session, channelID, content string, author *discordgo.User) (*database.Record, error) {
	resultErr := errors.New("parse failure")

	args := strings.Split(strings.TrimPrefix(content, "-запрос "), "->")
	if len(args) > 2 {
		s.ChannelMessageSend(channelID, "Обнаружена лишняя \"->\"")
		return nil, resultErr
	}
	if len(args) < 2 {
		s.ChannelMessageSend(channelID, "Где \"->\"?")
		return nil, resultErr
	}

	reason := strings.TrimSpace(args[0])
	if len(reason) == 0 {
		s.ChannelMessageSend(channelID, "Причина не должна быть пустой")
		return nil, resultErr
	}

	item := &database.Record{}
	item.Reason = reason
	item.AuthorName = author.Username
	item.AuthorIcon = author.AvatarURL("128")

	if len(args[1]) == 0 {
		s.ChannelMessageSend(channelID, "Не указаны юзвери и их деньги")
		return nil, resultErr
	}

	args[1] = makeParsingBetter(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(args[1]), ","), ","))

	var usersSumDB database.Uss
	pairsUsersSum := strings.Split(args[1], ",")
	for _, pairUsersSum := range pairsUsersSum {
		if len(pairUsersSum) == 0 {
			s.ChannelMessageSend(channelID, "Не указаны юзвери и их деньги")
			return nil, resultErr
		}

		usersSum, err := splitReverse1(strings.TrimSpace(pairUsersSum), " ")

		if len(usersSum[0]) == 0 {
			s.ChannelMessageSend(channelID, "В указании юзверей или суммы содержится ошибка")
			return nil, resultErr
		}

		sum, err := strconv.ParseUint(strings.TrimSpace(usersSum[1]), 10, 64)
		if err != nil {
			s.ChannelMessageSend(channelID, "В указании юзверей или суммы содержится ошибка")
			return nil, resultErr
		}

		users := strings.Split(usersSum[0], " ")

		for i := 0; i < len(users); i++ {
			if !strings.HasPrefix(users[i], "<@") || !strings.HasSuffix(users[i], ">") {
				s.ChannelMessageSend(channelID, "В юзверях затесался шпион")
				return nil, resultErr
			}
			users[i] = strings.TrimPrefix(users[i], "<@")
			users[i] = strings.TrimPrefix(users[i], "!")
			users[i] = strings.TrimSuffix(users[i], ">")

			_, err = s.GuildMember(config.GdHouseID, users[i])
			if err != nil {
				s.ChannelMessageSend(channelID, "В юзверях затесался шпион")
				return nil, resultErr
			}
		}

		usersSumDB.Us = append(usersSumDB.Us, database.Us{
			Users: users,
			Sum:   sum,
		})
	}

	usersSumJSON, err := json.Marshal(usersSumDB)
	if err != nil {
		fmt.Println("ERROR marshling to JSON:", err)
		return nil, resultErr
	}

	item.UsersSum = string(usersSumJSON)

	return item, nil
}

func resendRequest(s *discordgo.Session, channelID, content string, item *database.Record) error {
	resultErr := errors.New("resending failure")

	message, err := s.ChannelMessageSendEmbed(config.ChForRequestID, createEmbed(s, item, 225225))
	if err != nil {
		s.ChannelMessageSend(channelID, "Не удалось отправить запрос")
		fmt.Println("ERROR sending request", err)
		return resultErr
	}
	item.EmbedID = message.ID

	err = s.MessageReactionAdd(config.ChForRequestID, item.EmbedID, "✅")
	if err != nil {
		fmt.Println("ERROR adding reaction ✅", err)
		return resultErr
	}

	err = s.MessageReactionAdd(config.ChForRequestID, item.EmbedID, "🇽")
	if err != nil {
		fmt.Println("ERROR adding reaction 🇽", err)
		return resultErr
	}

	return nil
}

func emojiOnRequest(s *discordgo.Session, r *discordgo.MessageReactionAdd, item *database.Record) {
	switch r.Emoji.Name {
	case "✅":
		_, err = s.ChannelMessageEditEmbed(config.ChForRequestID, item.EmbedID, createEmbed(s, item, 255255))
		if err != nil {
			fmt.Println("ERROR edit embed on ✅", err)
			return
		}

	case "🇽":
		_, err = s.ChannelMessageEditEmbed(config.ChForRequestID, item.EmbedID, createEmbed(s, item, 15158332))
		if err != nil {
			fmt.Println("ERROR edit embed on 🇽", err)
			return
		}
	default:
		return
	}

	err = database.Records.Delete(item.EmbedID)
	if err != nil {
		fmt.Println("ERROR delete record", err)
		return
	}
}

func checkRequest(str string) (bool, string) {
	if strings.HasPrefix(str, "-запрос") {
		return true, strings.TrimPrefix(str, "-запрос")
	}
	if strings.HasPrefix(str, "- запрос") {
		return true, strings.TrimPrefix(str, "- запрос")
	}
	if strings.HasPrefix(str, "-  запрос") {
		return true, strings.TrimPrefix(str, "-  запрос")
	}

	return false, ""
}
