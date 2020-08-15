package logic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/bwmarrin/discordgo"
)

func isRemind(m *discordgo.Message) bool {
	if m.Author.ID == config.UsRemindorID &&
		strings.Contains(m.Content, ">, Скоро будет ") {
		return true
	}

	return false
}

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
	if (m.Author.ID == config.UsLikeID || m.WebhookID == config.WhLikeID) &&
		len(m.Embeds) > 0 &&
		len(m.Embeds[0].Description) > 0 &&
		strings.HasPrefix(m.Embeds[0].Description, "Вы успешно лайкнули сервер.") &&
		m.Embeds[0].Author != nil {
		return true
	}

	return false
}

func splitReverse1(str string, sep string) ([2]string, error) {
	for i := len(str) - 1; i > 0; i-- {
		if string(str[i]) == sep {
			return [2]string{str[:i], str[i+1:]}, nil
		}
	}

	return [2]string{"", str}, errors.New("")
}

func hasRole(member *discordgo.Member, id string) bool {
	for _, role := range member.Roles {
		if role == id {
			return true
		}
	}
	return false
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

func secondsToString(begin string, secs int64) string {
	if secs < 0 {
		return fmt.Sprint(begin, " проспали!\n")
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
