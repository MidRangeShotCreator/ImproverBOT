package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var gBot *tgbotapi.BotAPI
var gToken string
var gChatId int64

var gUsersInChat Users

var gUsefulActivities = Activities{
	// Саморазвитие
	{"yoga", "Йога (15 минут)", 1},
	{"meditation", "Медитация (15 минут)", 1},
	{"language", "Изучение английского (15 минут)", 1},
	{"swimming", "Танцы (15 минут)", 1},
	{"walk", "Прогулка (15 минут)", 1},
	{"chores", "Уборка", 1},

	// Работа
	{"work_learning", "Работа (15 минут)", 1},
	{"portfolio_work", "Создание pet-проекта (15 минут)", 1},
	{"resume_edit", "Редактирование резюме (15 минут)", 1},

	// Креативность
	{"creative", "Творческое созидание (15 минут)", 1},
	{"reading", "Чтение (15 минут)", 1},
}

var gRewards = Activities{
	// Entertainment
	{"watch_series", "Просмотр сериала (1 серия)", 10},
	{"watch_movie", "Просмотр фильма (1 шт)", 30},
	{"social_nets", "Просмотр соцсетей (30 минут)", 10},

	// Food
	{"eat_sweets", "300 ккал вкусняшек", 60},
}

type User struct {
	id    int64
	name  string
	coins uint16
}
type Users []*User

type Activity struct {
	code, name string
	coins      uint16
}
type Activities []*Activity

func init() {
	// Uncomment and update token value to set environment variable for Telegram Bot Token given by BotFather.
	// Delete this line after setting the env var. Keep the token out of the public domain!
	_ = os.Setenv(TOKEN_NAME_IN_OS, "6840789546:AAHD9rMEMoEs13-Bm4zexkGmkQWubdir8gc")
	
	if gToken = os.Getenv(TOKEN_NAME_IN_OS); gToken == "" {
		panic(fmt.Errorf(`failed to load environment variable "%s"`, TOKEN_NAME_IN_OS))
	}

	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	gBot.Debug = true
}

func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func isCallbackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

func delay(seconds uint8) {
	time.Sleep(time.Second * time.Duration(seconds))
}

func sendStringMessage(msg string) {
	gBot.Send(tgbotapi.NewMessage(gChatId, msg))
}

func sendMessageWithDelay(delayInSec uint8, message string) {
	sendStringMessage(message)
	delay(delayInSec)
}

func printIntro(update *tgbotapi.Update) {
	sendMessageWithDelay(2, "Привет! "+EMOJI_SUNGLASSES)
	sendMessageWithDelay(7, "Есть множество полезных действий, совершая которые на регулярной основе мы улучшаем качество своей жизни. Но часто гораздо веселее, проще или вкуснее сделать что-то вредное. Не так ли?")
	sendMessageWithDelay(7, "С большей вероятностью мы предпочтем залипнуть в YouTube Shorts вместо урока английмкого, купим M&M's вместо овощей, полежим на кровати вместо йоги.")
	sendMessageWithDelay(1, EMOJI_SAD)
	sendMessageWithDelay(10, "Каждый играл хотя бы в одну игру, где нужно прокачивать персонажа, делая его сильнее, умнее или красивее. Делать это приятно, потому что каждое действие приносит результаты. В реальной же жизни только систематические действия через время начинают быть заметны. Давай это изменим?")
	sendMessageWithDelay(1, EMOJI_SMILE)
	sendMessageWithDelay(14, `Перед тобой две таблицы: "Полезные действия" и "Вознаграждения". В первой таблице перечислены несложные короткие активности, за выполнение каждой из которых ты получишь указанное количество монет. Во второй таблице ты увидишь перечень активностей, сделать которые ты можешь только после того, как оплатишь их заработанными на предыдущем этапе монетами.`)
	sendMessageWithDelay(1, EMOJI_COIN)
	sendMessageWithDelay(10, `Например, ты пол часа занимаешься йогой, за что получаешь 2 монеты. После этого у тебя есть 2 часа изучения программирования, за что ты получаешь 8 монет. Теперь ты можешь посмотреть 1 серию "Не родись красивой" и выйти в ноль. Всё просто!`)
	sendMessageWithDelay(6, `Отмечай совершенные полезные активности, чтобы не потерять монеты. И не забывай "купить" вознаграждение перед тем как его совершить.`)
}

func getKeyboardRow(buttonText, buttonCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonCode))
}

func askToPrintIntro() {
	msg := tgbotapi.NewMessage(gChatId, "Во вступительных сообщениях ты можешь найти смысл данного бота и правила игры. Что думаешь?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_PRINT_INTRO, BUTTON_CODE_PRINT_INTRO),
		getKeyboardRow(BUTTON_TEXT_SKIP_INTRO, BUTTON_CODE_SKIP_INTRO),
	)
	gBot.Send(msg)
}

func showMenu() {
	msg := tgbotapi.NewMessage(gChatId, "Выбери один из вариантов:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_BALANCE, BUTTON_CODE_BALANCE),
		getKeyboardRow(BUTTON_TEXT_USEFUL_ACTIVITIES, BUTTON_CODE_USEFUL_ACTIVITIES),
		getKeyboardRow(BUTTON_TEXT_REWARDS, BUTTON_CODE_REWARDS),
	)
	gBot.Send(msg)
}

func showBalance(user *User) {
	msg := fmt.Sprintf("%s, твой кошелёк пока пуст %s \nЗатрекай полезное действие, чтобы получить монеты.", user.name, EMOJI_DONT_KNOW)
	if coins := user.coins; coins > 0 {
		msg = fmt.Sprintf("%s, у тебя %d %s", user.name, coins, EMOJI_COIN)
	}
	sendStringMessage(msg)
	showMenu()
}

func callbackQueryFromIsMissing(update *tgbotapi.Update) bool {
	return update.CallbackQuery == nil || update.CallbackQuery.From == nil
}

func getUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryFromIsMissing(update) {
		return
	}

	userId := update.CallbackQuery.From.ID
	for _, userInChat := range gUsersInChat {
		if userId == userInChat.id {
			return userInChat, true
		}
	}
	return
}

func storeUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryFromIsMissing(update) {
		return
	}

	from := update.CallbackQuery.From
	user = &User{id: from.ID, name: strings.TrimSpace(from.FirstName + " " + from.LastName), coins: 0}
	gUsersInChat = append(gUsersInChat, user)
	return user, true
}

func showActivities(activities Activities, message string, isUseful bool) {
	activitiesButtonsRows := make([]([]tgbotapi.InlineKeyboardButton), 0, len(activities)+1)
	for _, activity := range activities {
		activityDescription := ""
		if isUseful {
			activityDescription = fmt.Sprintf("+ %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		} else {
			activityDescription = fmt.Sprintf("- %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		}
		activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(activityDescription, activity.code))
	}
	activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(BUTTON_TEXT_PRINT_MENU, BUTTON_CODE_PRINT_MENU))

	msg := tgbotapi.NewMessage(gChatId, message)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(activitiesButtonsRows...)
	gBot.Send(msg)
}

func showUsefulActivities() {
	showActivities(gUsefulActivities, "Трекай полезное действие или возвращайся в главное меню:", true)
}

func showRewards() {
	showActivities(gRewards, "Купи вознаграждение или возвращайся в главное меню:", false)
}

func findActivity(activities Activities, choiceCode string) (activity *Activity, found bool) {
	for _, activity := range activities {
		if choiceCode == activity.code {
			return activity, true
		}
	}
	return
}

func processUsefulActivity(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`у активности "%s" не указана стоимость`, activity.name)
	} else if user.coins+activity.coins > MAX_USER_COINS {
		errorMsg = fmt.Sprintf("у тебя не может быть больше %d %s", MAX_USER_COINS, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, прости, но %s %s Твой баланс остался без изменений.", user.name, errorMsg, EMOJI_SAD)
	} else {
		user.coins += activity.coins
		resultMessage = fmt.Sprintf(`%s, действие "%s" выполнено %d %s Поступило к тебе на счёт. Так держать! %s%s Теперь у тебя %d %s`,
			user.name, activity.name, activity.coins, EMOJI_COIN, EMOJI_BICEPS, EMOJI_SUNGLASSES, user.coins, EMOJI_COIN)
	}
	sendStringMessage(resultMessage)
}

func processReward(activity *Activity, user *User) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`у вознаграждения "%s" не указана стоимость`, activity.name)
	} else if user.coins < activity.coins {
		errorMsg = fmt.Sprintf(`у тебя сейчас %d %s. Ты не можешь позволить себе "%s" за %d %s`, user.coins, EMOJI_COIN, activity.name, activity.coins, EMOJI_COIN)
	}

	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, прости, но %s %s Твой баланс остался без изменений, вознаграждение недоступно %s", user.name, errorMsg, EMOJI_SAD, EMOJI_DONT_KNOW)
	} else {
		user.coins -= activity.coins
		resultMessage = fmt.Sprintf(`%s, вознаграждение "%s" оплачено, приступай! %d %s было снято с твоего счёта. Теперь у тебя %d %s`, user.name, activity.name, activity.coins, EMOJI_COIN, user.coins, EMOJI_COIN)
	}
	sendStringMessage(resultMessage)
}

func updateProcessing(update *tgbotapi.Update) {
	user, found := getUserFromUpdate(update)
	if !found {
		if user, found = storeUserFromUpdate(update); !found {
			sendStringMessage("Не получается идентифицировать пользвателя")
			return
		}
	}

	choiceCode := update.CallbackQuery.Data
	log.Printf("[%T] %s", time.Now(), choiceCode)

	switch choiceCode {
	case BUTTON_CODE_BALANCE:
		showBalance(user)
	case BUTTON_CODE_USEFUL_ACTIVITIES:
		showUsefulActivities()
	case BUTTON_CODE_REWARDS:
		showRewards()
	case BUTTON_CODE_PRINT_INTRO:
		printIntro(update)
		showMenu()
	case BUTTON_CODE_SKIP_INTRO:
		showMenu()
	case BUTTON_CODE_PRINT_MENU:
		showMenu()
	default:
		if usefulActivity, found := findActivity(gUsefulActivities, choiceCode); found {
			processUsefulActivity(usefulActivity, user)

			delay(2)
			showUsefulActivities()
			return
		}

		if reward, found := findActivity(gRewards, choiceCode); found {
			processReward(reward, user)

			delay(2)
			showRewards()
			return
		}

		log.Printf(`[%T] !!!!!!!!! ERROR: Unknown code "%s"`, time.Now(), choiceCode)
		msg := fmt.Sprintf("%s, Прости, я не знаю код '%s' %s Пожалуйста, сообщи моему создателю об ошибке.", user.name, choiceCode, EMOJI_SAD)
		sendStringMessage(msg)
	}
}

func main() {
	log.Printf("Authorized on account %s", gBot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	for update := range gBot.GetUpdatesChan(updateConfig) {
		if isCallbackQuery(&update) {
			updateProcessing(&update)
		} else if isStartMessage(&update) {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			gChatId = update.Message.Chat.ID
			askToPrintIntro()
		}
	}
}