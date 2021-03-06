package config

import (
	"os"
	"strconv"
	"time"

	"github.com/SteMak/ani-helper/workerTools/bankirapi"
)

// Usage var of basic syntax
const Usage = "Общий синтаксис:\n\t-запрос <причина> -> <заслужившие награду пользователи> <сумма для выдачи>, <заслужившие другую награду пользователи> <сумма для выдачи>\n" +
	"Пример:\n\t-запрос Победители ивента \"Шляпа\" -> @someone @elseone 100, @anotherone 200"

var (
	// Token discord token
	Token string
	// BankirapiToken bankirapi token
	BankirapiToken string

	// API an interface for bankir bot
	API *bankirapi.API

	// GdHouseID guildID
	GdHouseID string

	// SumForPaying money for bump and s.up
	SumForPaying string

	// ChForRequestID channelID
	ChForRequestID string
	// ChForLogsID channelID
	ChForLogsID string
	// ChForBustsID for monitoring
	ChForBustsID string
	// ChForInfoRequestID for give info about award to user
	ChForInfoRequestID string

	// RoConfirmatorID userID
	RoConfirmatorID string
	// UsSiupID siup user
	UsSiupID string
	// UsBumpID bump user
	UsBumpID string
	// UsLikeID like user
	UsLikeID string
	// UsRemindorID me user
	UsRemindorID string
	// UsMainWorker roleID
	UsMainWorker string

	// WhLikeID me user
	WhLikeID string

	// TimeWaitSiup siup time
	TimeWaitSiup int
	// TimeWaitBump bump time
	TimeWaitBump int
	// TimeWaitLike like time
	TimeWaitLike int
	// TimeRemind remind time
	TimeRemind int
	// TimeDoubleRemind remind time
	TimeDoubleRemind int

	// LastSiup siup user
	LastSiup time.Time
	// LastBump bump user
	LastBump time.Time
	// LastLike like user
	LastLike time.Time
	// LastRemind remind user
	LastRemind time.Time

	// RoRequestMakerID roleID
	RoRequestMakerID string
	// RoBuster roleID
	RoBuster string

	// PostgresURI uri of postgress
	PostgresURI string

	// Responces beautiful log
	Responces = []string{
		"Вы сделали %s сервера, и Тихий Ужас вручил Вам %s",
		"Вы героически поймали %s, и Вас наградили %s",
		"за %s сервера Жмяк отдала Вам свои печеньки, и Вы получили %s",
		"после кибератаки Вы подняли сервер своим %sом и получили %s",
		"Вы, пожертвовав жизнью на войне за %s, были посмертно награждены %s",
		"смакуючи ельфійською абракадаброю (%s), Ви начарували %s",
		"сыр съел сырный сырник, %sая сервер, Вам заплатили моральную компенсацию в %s",
		"Вы помогли Ведьмаку с %sом, за что Вам заплатили __ЧЕКАННЫМИ__ %s",
		"Вы сделали %s сервера, и Скромный Модератор вручил Вам %s",
		"Вы съели свою поджелудочную во время %sа и нашли в ней %s",
		"Вы собрали с тысячи людей по АниКоину и %sнули сервер. Вы получили %s",
		"Вы попытались выписать бан Кнопычу, но сделали %s и получили %s",
		"на вечерних посиделках с Хикаро Вы сражались за %s. Хикаро Вас наградил %s",
		"Вы сохранили последние пять минут и угрожающе сделали %s. Глюк растрогался и отдал Вам %s",
		"Вы наблюдали за программированием Стёмы и Меро, но не забыли сделать %s и получили %s",
		"Меро уходил спать, но %sнул за Вас, и Вы получили %s",
		"Нев отвлёк всех разговорами об отсутствии холодильника, и Вы %sнули сервер. Холодильник дал вам %s",
		"Кемпер поднимал актив, чтобы кемперить АниКоины, так что Вы тихо %sнули сервер и получили %s",
		"Вы вычислили Маргинала, и он рассказал Вам секрет %sанья. Получено %s",
		"у Боннуса провис интернет, так что вы беспрепятственно %sнули сервер и забрали %s",
		"Эксля заснул, что-то бормоча во сне: \"Z-z-z... %s Z-z-z... %s Z-z-z...\"",
		"Эспада зажала Вас в тиски, но вы успели сделать %s. Они зауважали Вас и дали %s",
		"Маю-Маю снова уснула в войсе, Вы вдохновились её ворочанием и %sнули, заработав %s",
		"ɔıloqoou написал %s перевёрнутыми буквами, поэтому Вы неспеша забрали %s",
		"Фузу мирно рисовала в войсе, а Вы сделали %s и собрали %s",
		"Кнопка рассматривал ответы бота, Вы, сделав %s, добавили ещё один, и недоумевающий Кнопыч вручил Вам %s",
	}

	// FairyReplacement for good parse
	FairyReplacement = [][2]string{
		{"  ", " "},
		{",,", ","},
		{", ,", ","},
		{"><", "> <"},
		{">1", "> 1"},
		{">2", "> 2"},
		{">3", "> 3"},
		{">4", "> 4"},
		{">5", "> 5"},
		{">6", "> 6"},
		{">7", "> 7"},
		{">8", "> 8"},
		{">9", "> 9"},
		{">0", "> 0"},
	}
)

// Init inits main vars for the project
func Init() {
	Token = os.Getenv("TOKEN")
	BankirapiToken = os.Getenv("BANKIRAPI_TOKEN")

	API = bankirapi.New(BankirapiToken)

	GdHouseID = os.Getenv("GD_HOUSE_ID")

	SumForPaying = os.Getenv("SUM_FOR_PAYING")

	ChForRequestID = os.Getenv("CH_FOR_REQUEST_ID")
	ChForLogsID = os.Getenv("CH_FOR_LOGS_ID")
	ChForBustsID = os.Getenv("CH_FOR_BUSTS_ID")
	ChForInfoRequestID = os.Getenv("CH_FOR_INFO_REQUEST_ID")

	RoConfirmatorID = os.Getenv("RO_CONFIRMATOR_ID")
	UsSiupID = os.Getenv("US_SIUP_ID")
	UsBumpID = os.Getenv("US_BUMP_ID")
	UsLikeID = os.Getenv("US_LIKE_ID")
	UsRemindorID = os.Getenv("US_REMINDOR_ID")
	UsMainWorker = os.Getenv("US_MAIN_WORKER")

	WhLikeID = os.Getenv("WH_LIKE_ID")

	TimeWaitSiup, _ = strconv.Atoi(os.Getenv("TIME_WAIT_SIUP"))
	TimeWaitBump, _ = strconv.Atoi(os.Getenv("TIME_WAIT_BUMP"))
	TimeWaitLike, _ = strconv.Atoi(os.Getenv("TIME_WAIT_LIKE"))
	TimeRemind, _ = strconv.Atoi(os.Getenv("TIME_REMIND"))
	TimeDoubleRemind, _ = strconv.Atoi(os.Getenv("TIME_DOUBLE_REMIND"))

	RoRequestMakerID = os.Getenv("RO_REQUEST_MAKER_ID")
	RoBuster = os.Getenv("RO_BUSTER")

	PostgresURI = os.Getenv("DATABASE_URL")
}
