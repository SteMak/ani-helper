package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/SteMak/ani-helper/workerTools/database"
	"github.com/bwmarrin/discordgo"
)

var (
	err error
)

func isSiup(m *discordgo.Message) bool {
	if m.Author.ID == config.UsSiupID &&
		len(m.Embeds) > 0 &&
		m.Embeds[0].Title == "Сервер Up" &&
		m.Embeds[0].Footer != nil {
		return true
	}

	return false
}

func isBump(m *discordgo.Message) bool {
	if m.Author.ID == config.UsBumpID &&
		len(m.Embeds) > 0 {

		matched, err := regexp.Match(`Server bumped by <@\d+>`, []byte(m.Embeds[0].Description))
		if err != nil {
			fmt.Println("ERROR Bump make match regular failure:", err)
			return false
		}
		if matched {
			return true
		}
	}

	return false
}

func isLike(m *discordgo.Message) bool {
	if m.Author.ID == config.UsLikeID &&
		len(m.Embeds) > 0 &&
		len(m.Embeds[0].Description) > 0 &&
		m.Embeds[0].Description[:50] == "Вы успешно лайкнули сервер." &&
		m.Embeds[0].Author != nil {
		return true
	}

	return false
}

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
	fmt.Println("FOUND S.up")

	for _, user := range chMonitorWriters {
		if user.strify == m.Embeds[0].Footer.Text {
			fmt.Println("FOUND S.up user", user.id)

			sendAndLog(s, user.id, "S.up", config.SumForPaying)
			go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind))
			return
		}
	}
}

func onBumpServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.LastBump = time.Now()
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
	go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind))
}

func onLikeServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	config.LastLike = time.Now()
	fmt.Println("FOUND Like")

	for _, user := range chMonitorWriters {
		if user.strify == m.Embeds[0].Author.Name {
			fmt.Println("FOUND Like user", user.id)

			sendAndLog(s, user.id, "Like", config.SumForPaying)
			go sleep(s, int64(config.TimeWaitBump*60-config.TimeRemind))
			return
		}
	}
}

func sendAndLog(s *discordgo.Session, userID string, str string, sum string) {
	isum, err := strconv.Atoi(sum)
	if err != nil {
		fmt.Println("ERROR "+str+" parsing sum:", err)
	}
	_, err = config.API.AddToBalance(config.GdHouseID, userID, 0, isum, "for "+str)
	if err != nil {
		fmt.Println("ERROR "+str+" updating user balance:", err)

		_, err = s.ChannelMessageSend(config.ChForLogsID, "Кажись, что-то пошло не так... <@"+userID+"> сделал "+str+", но денег ему не дали(")
		if err != nil {
			fmt.Println("ERROR "+str+" sending wrong report message:", err)
		}
		_, err = s.ChannelMessageSend(config.ChForBustsID, "<@"+userID+">, у нас снова что-то сломалось, но не волнуйтесь - деньги Вам прилетят чуть позже)")
		if err != nil {
			fmt.Println("ERROR "+str+" sending wrong log message:", err)
		}

		return
	}

	_, err = s.ChannelMessageSend(config.ChForLogsID, sum+"<:AH_AniCoin:579712087224483850> были выданы <@"+userID+">, за то что он сделал "+str)
	if err != nil {
		fmt.Println("ERROR "+str+" sending right report message:", err)
	}

	_, err = s.ChannelMessageSend(config.ChForBustsID, "<@"+userID+">, "+fmt.Sprintf(config.Responces[rand.Intn(len(config.Responces))], str, sum+"<:AH_AniCoin:579712087224483850>"))
	if err != nil {
		fmt.Println("ERROR "+str+" sending right log message:", err)
	}

	fmt.Println("GUILD "+str+" by  ", userID)
}

func splitReverse1(str string, sep string) ([2]string, error) {
	for i := len(str) - 1; i > 0; i-- {
		if string(str[i]) == sep {
			return [2]string{str[:i], str[i+1:]}, nil
		}
	}

	return [2]string{"", str}, errors.New("")
}

func createEmbed(s *discordgo.Session, item *database.Record, color int) *discordgo.MessageEmbed {
	var users string

	uss := database.Uss{}
	json.Unmarshal([]byte(item.UsersSum), &uss)
	for _, userSum := range uss.Us {
		users += strconv.FormatUint(userSum.Sum, 10) + "\n<@" + strings.Join(userSum.Users, "> <@") + ">\n"

		if color == 255255 {
			for _, user := range userSum.Users {
				_, err := config.API.AddToBalance(config.GdHouseID, user, 0, int(userSum.Sum), item.Reason)
				if err != nil {
					fmt.Println("ERROR "+item.EmbedID+" add money to user balance:", err)
					_, err = s.ChannelMessageSend(config.ChForLogsID, "Кажись, что-то пошло не так... <@"+user+"> не получил денег за "+item.Reason)
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending wrong report message:", err)
					}
				} else {
					_, err = s.ChannelMessageSend(config.ChForInfoRequestID, "<@"+user+">, "+"Вы получили "+strconv.FormatUint(userSum.Sum, 10)+"<:AH_AniCoin:579712087224483850>. Причина: "+item.Reason)
					if err != nil {
						fmt.Println("ERROR "+item.EmbedID+" sending right info message:", err)
					}
					_, err = s.ChannelMessageSend(config.ChForLogsID, strconv.FormatUint(userSum.Sum, 10)+"<:AH_AniCoin:579712087224483850> были выданы <@"+user+">, за "+item.Reason)
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

func hasRole(member *discordgo.Member, id string) bool {
	for _, role := range member.Roles {
		if role == id {
			return true
		}
	}
	return false
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

func makeParsingBetter(str string) string {
	result := str
	for _, rep := range config.FairyReplacement {
		result = strings.Join(strings.Split(result, rep[0]), rep[1])
	}

	if result != str {
		return makeParsingBetter(result)
	}

	return result
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

func isRemind(m *discordgo.Message) bool {
	if m.Author.ID == config.UsRemindorID &&
		strings.Contains(m.Content, ">, Скоро будет ") {
		return true
	}

	return false
}

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

func checkAndRemind(s *discordgo.Session) {
	if time.Now().Unix() >= config.LastLike.Unix()+int64(config.TimeWaitLike*60-config.TimeRemind) {
		remind(s, "Like")
	}
	if time.Now().Unix() >= config.LastBump.Unix()+int64(config.TimeWaitBump*60-config.TimeRemind) {
		remind(s, "Bump")
	}
	if time.Now().Unix() >= config.LastSiup.Unix()+int64(config.TimeWaitSiup*60-config.TimeRemind) {
		remind(s, "Siup")
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
	time.Sleep(time.Duration(config.TimeRemind) * time.Second)

	_, err := s.ChannelMessageSend(config.ChForBustsID, str+" заряжен!")
	if err != nil {
		fmt.Println("ERROR LastRemind send:", err)
		return
	}
	fmt.Println("PRINT LastRemind " + str)
}

func secondsToString(begin string, secs int64) string {
	if secs < 0 {
		return fmt.Sprint(begin, " проспали!")
	}
	text := ""
	ho := ""
	mi := ""
	se := ""
	h := int(secs / 3600)
	if h-h/10*10 > 4 || h-h/10*10 == 0 || h-h/100*100 > 10 && h-h/100*100 < 15 {
		ho = " часов "
	} else if h-h/10*10 == 1 {
		ho = " час "
	} else {
		ho = " часа "
	}
	m := int(secs/60) - h*60
	if m-m/10*10 > 4 || m-m/10*10 == 0 || m-m/100*100 > 10 && m-m/100*100 < 15 {
		mi = " минут "
	} else if m-m/10*10 == 1 {
		mi = " минуту "
	} else {
		mi = " минуты "
	}
	s := int(secs) - h*3600 - m*60
	if s-s/10*10 > 4 || s-s/10*10 == 0 || s-s/100*100 > 10 && s-s/100*100 < 15 {
		se = " секунд "
	} else if s-s/10*10 == 1 {
		se = " секунду "
	} else {
		se = " секунды "
	}
	if h > 0 {
		text = fmt.Sprint(begin, " будет через ", h, ho, m, mi, "и ", s, se, " \n")
	} else if m > 0 {
		text = fmt.Sprint(begin, " будет через ", m, mi, "и ", s, se, " \n")
	} else {
		text = fmt.Sprint(begin, " будет через ", s, se, " \n")
	}
	return text
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
