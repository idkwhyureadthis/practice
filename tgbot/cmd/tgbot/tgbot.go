package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/idkwhyureadthis/tgbot/internal/models"
)

var TOKEN = os.Getenv("TOKEN")

var bot *tgbotapi.BotAPI

type DBFilter struct {
	name       string
	city       string
	salaryFrom int
	experience int
}

type ParseFilter struct {
	name       string
	salaryFrom int
}

type StateMachine struct {
	State string
}

func (s *StateMachine) ChangeState(newState string) {
	s.State = newState
}

var parseFilters = make(map[int64]ParseFilter)
var dbFilters = make(map[int64]DBFilter)
var states = make(map[int64]StateMachine)

func connectWithTelegram() {
	var err error
	if bot, err = tgbotapi.NewBotAPI(TOKEN); err != nil {
		log.Fatal("Failed to connect to Telegram")
	}
}

func main() {
	var PS = fmt.Sprintf("%v", os.PathSeparator)
	var LineBreak = "\n"
	if PS != "/" {
		LineBreak = "\r\n"
	}

	connectWithTelegram()
	log.Println("bot succesfully started")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Text == "/start" {
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Парсинг вакансий", "parse_vacancies"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Поиск в БД", "get_from_db"),
					),
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
				msg.ReplyMarkup = inlineKeyboard
				states[update.Message.Chat.ID] = StateMachine{State: "SelectingAction"}
				bot.Send(msg)
			}
			if states[update.Message.Chat.ID].State == "ParsingVacanciesSalary" {
				minSalary, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					minSalary = 0
				}
				parseFilters[update.Message.Chat.ID] = ParseFilter{
					name:       parseFilters[update.Message.Chat.ID].name,
					salaryFrom: minSalary,
				}
				pf := parseFilters[update.Message.Chat.ID]
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("паршу вакансии с названием %s и минимальной зарплатой %d рублей", pf.name, pf.salaryFrom))
				bot.Send(msg)
				u, _ := url.Parse("http://localhost:8080/parse")
				q := u.Query()
				q.Add("text", pf.name)
				q.Add("min_salary", fmt.Sprint(pf.salaryFrom))
				u.RawQuery = q.Encode()
				resp, _ := http.Get(u.String())
				if resp.StatusCode != 200 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не удалось получить данные с сервера")
					bot.Send(msg)
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Парсинг вакансий", "parse_vacancies"),
						),
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Поиск в БД", "get_from_db"),
						),
					)

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
					msg.ReplyMarkup = inlineKeyboard
					states[update.Message.Chat.ID] = StateMachine{State: "SelectingAction"}
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Данные были успешно записаны")
					bot.Send(msg)
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Парсинг вакансий", "parse_vacancies"),
						),
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Поиск в БД", "get_from_db"),
						),
					)

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
					msg.ReplyMarkup = inlineKeyboard
					states[update.Message.Chat.ID] = StateMachine{State: "SelectingAction"}
					bot.Send(msg)
				}
			}
			if states[update.Message.Chat.ID].State == "ParsingVacanciesName" {
				parseFilters[update.Message.Chat.ID] = ParseFilter{
					name: update.Message.Text,
				}
				states[update.Message.Chat.ID] = StateMachine{State: "ParsingVacanciesSalary"}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите минимальную зарплату в рублях для парсинга (можете написать 0 для парсинга любых вакансий)")
				bot.Send(msg)
			}
			if states[update.Message.Chat.ID].State == "GettingVacanciesExperience" {
				experience, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					experience = 0
				}
				dbFilters[update.Message.Chat.ID] = DBFilter{
					name:       dbFilters[update.Message.Chat.ID].name,
					salaryFrom: dbFilters[update.Message.Chat.ID].salaryFrom,
					city:       dbFilters[update.Message.Chat.ID].city,
					experience: experience,
				}
				filters := dbFilters[update.Message.Chat.ID]
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ищу вакансии по данному запросу...")
				bot.Send(msg)

				u, _ := url.Parse("http://localhost:8080/get")
				q := u.Query()
				q.Add("name", filters.name)
				q.Add("city", filters.city)
				q.Add("min_salary", fmt.Sprint(filters.salaryFrom))
				q.Add("experience", fmt.Sprint(filters.experience))
				u.RawQuery = q.Encode()

				resp, _ := http.Get(u.String())
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не удалось получить данные из базы данных")
					bot.Send(msg)
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Парсинг вакансий", "parse_vacancies"),
						),
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Поиск в БД", "get_from_db"),
						),
					)

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
					msg.ReplyMarkup = inlineKeyboard
					states[update.Message.Chat.ID] = StateMachine{State: "SelectingAction"}
					bot.Send(msg)
				} else {
					body, _ := io.ReadAll(resp.Body)
					var data models.PageData
					_ = json.Unmarshal(body, &data)
					file, _ := json.MarshalIndent(data, "", "")
					_ = os.WriteFile("resp.json", file, 0644)
					doc := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath("resp.json"))
					bot.Send(doc)
					err = os.Remove("resp.json")
					if err != nil {
						log.Println(err)
					}
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Парсинг вакансий", "parse_vacancies"),
						),
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Поиск в БД", "get_from_db"),
						),
					)

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")
					msg.ReplyMarkup = inlineKeyboard
					states[update.Message.Chat.ID] = StateMachine{State: "SelectingAction"}
					bot.Send(msg)
				}
			}
			if states[update.Message.Chat.ID].State == "GettingVacanciesCity" {
				dbFilters[update.Message.Chat.ID] = DBFilter{
					name:       dbFilters[update.Message.Chat.ID].name,
					salaryFrom: dbFilters[update.Message.Chat.ID].salaryFrom,
					city:       update.Message.Text,
				}
				states[update.Message.Chat.ID] = StateMachine{State: "GettingVacanciesExperience"}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваш опыт работы в формате числа от 0 до 3"+LineBreak+"0 - нет опыта"+LineBreak+"1 - от 1 года до 3 лет"+LineBreak+"2 - от 3 лет до 6 лет"+LineBreak+"3 - более 6 лет")
				bot.Send(msg)
			}
			if states[update.Message.Chat.ID].State == "GettingVacanciesSalary" {

				minSalary, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					minSalary = 0
				}
				dbFilters[update.Message.Chat.ID] = DBFilter{
					name:       dbFilters[update.Message.Chat.ID].name,
					salaryFrom: minSalary,
				}
				states[update.Message.Chat.ID] = StateMachine{State: "GettingVacanciesCity"}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите название города")
				bot.Send(msg)
			}
			if states[update.Message.Chat.ID].State == "GettingVacanciesName" {
				dbFilters[update.Message.Chat.ID] = DBFilter{
					name: update.Message.Text,
				}
				states[update.Message.Chat.ID] = StateMachine{State: "GettingVacanciesSalary"}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите минимальную зарплату в рублях для получения из бд (можете написать 0 для парсинга любых вакансий)")
				bot.Send(msg)
			}

		} else if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "parse_vacancies" {
				states[update.CallbackQuery.Message.Chat.ID] = StateMachine{"ParsingVacanciesName"}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите название вакансии для парсинга")
				bot.Send(msg)
			}
			if update.CallbackQuery.Data == "get_from_db" {
				states[update.CallbackQuery.Message.Chat.ID] = StateMachine{"GettingVacanciesName"}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите название вакансии для получения из базы данных")
				bot.Send(msg)
			}
		}
	}
}
