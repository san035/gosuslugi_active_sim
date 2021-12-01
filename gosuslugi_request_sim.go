/*
	Подтверждение сим карт Мегафон на сайте gosuslugi.ru в ЛК физлица по ФЗ-533

 В текущей папке должен быть:
1) chromedriver.exe скаченный с https://chromedriver.chromium.org/downloads выбор версии зависит от вашей версии браузера Chrome
2) config.ini c параметрами
[gosuslugi.ru]
login=000-111-222 33
password=4654Jhytgs

Исходники:
https://github.com/san035/gosuslugi_request_sim

Команда компиляции exe фала:
go build active_sim.go
*/

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"gopkg.in/ini.v1"
)

var driver_service *selenium.Service
var web_driver selenium.WebDriver
var err error
var max_wait_sec, skip_message int
var last_web_element selenium.WebElement
var last_web_elements []selenium.WebElement
var last_xpath string
var xpaths = map[string]string{
	"input_login":     `//input[@id="login"]`,
	"input_password":  `//input[@id="password"]`,
	"button_login":    `//button[text()="Войти"]`,
	"Войти_как":       `//p[contains(text(),"Частное лицо")]`,
	"Показать_больше": `//span[text()="Показать больше"]`,
	"Номер_телефона_в_сообщении": `//div/p/b`,
	"Кнопка_Проверить_данные":    `//a[contains(text(),"Проверить данные")]`,
	"Кнопка_Подтвердить_данные":  `//button/span[contains(text(),"Подтвердить")]`,
	"Заявление_отправлено":       `//h3/center[contains(text(), "Заявление отправлено")]`,
	"xpath_right": `//button/span[contains(text(),"Верно")]`,
}

func main() {
	//логирование на экран и в файл
	fileLog := prepare_log("sim.log")
	defer fileLog.Close()

	// Читаем config.ini
	cfg, err_cfg := ini.Load("config.ini")
	if err_cfg != nil {
		log.Fatalln("Error", "config.ini", err_cfg.Error())
	}

	// получаем логин/пароль от сайта
	login := cfg.Section("gosuslugi.ru").Key("login").String()
	if login == "" {
		login, _ = os.LookupEnv("gosuslugi_login")
	}
	password := cfg.Section("gosuslugi.ru").Key("password").String()
	if password == "" {
		password, _ = os.LookupEnv("gosuslugi_password")
	}
	log.Println("login, password:", login, strings.Repeat("*", len(password)))

	max_wait_sec, _ = cfg.Section("gosuslugi.ru").Key("max_wait_sec").Int()
	skip_message, _ = cfg.Section("gosuslugi.ru").Key("skip_message").Int()

	//загружаем xpath из файла
	for key_xpath, value_xpath := range xpaths {
		new_value := cfg.Section("xpath").Key(key_xpath).String()
		if new_value == "" {
			log.Println("Не найдено значение config.ini [xpath]"+key_xpath+" значение по умолчанию "+value_xpath, login)
		} else {
			xpaths[key_xpath] = new_value
		}
	}

	//Подготовка браузера к работе
	prepare_browser()
	defer driver_service.Stop()
	defer web_driver.Quit()

	//авторизация
	log.Println("Авторизация на сайте")
	open_url("https://lk.gosuslugi.ru/notifications?type=GEPS")
	time.Sleep(1 * time.Second)
	send_value("input_login", login)
	send_value("input_password", password)
	if login != "" && password != "" {
		press_button("button_login")
		err = nil //чтобы логин-пароль ввести вручную при ошибке
	}
	time.Sleep(2 * time.Second)
	press_button("Войти_как")

	//Открытие страницы с сообщениями
	log.Println("Поиск сообщений по сим")
	open_url("https://lk.gosuslugi.ru/notifications?type=GEPS")

	//Открываем все сообщения, нажатием //span[text()="Показать больше"]
	for find_web_element("Показать_больше") {
		last_web_element.Click()
		time.Sleep(2 * time.Second)
	}
	err = nil // т.к. последний поиск всегда неудачный

	//сбор сообщений для активаций сим
	find_web_element_array(`//a[contains(@href, "/message/")]//h4[text()="Запрос на активацию корпоративной сим-карты"]/../../../../..`)

	//сохранение url всех сообщений в hrefs_message
	var hrefs_message []string
	for _, last_web_element := range last_web_elements {
		url_message, err := last_web_element.GetAttribute("href")
		if err != nil {
			break
		}
		hrefs_message = append(hrefs_message, url_message)
	}
	log.Println("Найдено ", len(last_web_elements), `сообщений "Запрос на активацию корпоративной сим-карты"`) // , hrefs_message

	// обход всех сообщений
	arr_xpath := []string{xpaths["xpath_right"], `//h4[contains(text(), "Вы уже проверили данные")]`}
	for index, url_message := range hrefs_message {

		//Пропускаем первые skip_message сообщений
		if index < skip_message {
			continue
		}

		var add_info string
		open_url("https://lk.gosuslugi.ru" + url_message)

		// берем номер телефона
		if !find_web_element("Номер_телефона_в_сообщении") {
			add_info = " не найден телефон в сообщении"
			log.Println(index+1, url_message, add_info)
			err = nil
			continue
		}

		if err == nil && last_web_element != nil {
			add_info, err = last_web_element.Text()
		}

		press_button("Кнопка_Проверить_данные")

		//Поиск "Вы уже проверили данные" или "//button/span[contains(text(),"Верно")]"
		find_web_element_by_xpaths(arr_xpath)
		if err != nil {
			err = nil
			log.Println("Сообщение не проверено:", url_message)

		} else if last_xpath == xpaths["xpath_right"] {
			press_button("xpath_right")
			press_button("Кнопка_Подтвердить_данные")
			find_web_element("Заявление_отправлено")
			if err == nil {
				add_info += " отправлено сейчас"
			}
		} else {
			add_info += " отправлено ранее"
		}
		log.Println(index+1, add_info)
		if err != nil {
			break
		}

	}

	if err != nil {
		log.Println(err.Error())
	}

	log.Println("End ", err)
	time.Sleep(200 * time.Second)
}

// подключение к браузеру
func prepare_browser() { // (driver_service *selenium.Service) web_driver selenium.WebDriver
	log.Println("Открытие браузера chromedriver.exe")
	ops := []selenium.ServiceOption{}
	driver_service, err = selenium.NewChromeDriverService(`chromedriver.exe`, 9515, ops...)
	if err != nil {
		log.Fatalf("Error starting the ChromeDriver server: %v", err)
	}

	//Call browser
	//Set browser compatibility. We set the browser name to chrome
	caps := selenium.Capabilities{"browserName": "chrome"} // , "credentials_enable_service": false, "profile.password_manager_enabled": false
	web_driver, err = selenium.NewRemote(caps, "http://127.0.0.1:9515/wd/hub")
	if err != nil {
		panic(err)
	}
	return
}

func open_url(url string) {
	if err != nil {
		return
	}

	if err = web_driver.Get(url); err != nil {
		return
	}

}

//отправка значения value в поле ввода id
func send_value(xpath, value string) {
	if err != nil {
		return
	}

	find_web_element(xpath)
	if err != nil || value == "" {
		return
	}

	err = last_web_element.SendKeys(value)
}

//поиск xpath на странице, ожидая max_wait_sec секунд
func find_web_element(xpath string) bool {
	if err != nil {
		return false
	}

	if xpath[1:2] != "/" {
		xpath = xpaths[xpath]
	}

	for i := 0; i < max_wait_sec; i++ {
		last_xpath = xpath
		last_web_element, err = web_driver.FindElement(selenium.ByXPATH, xpath)
		if err == nil {
			return true
		}
		time.Sleep(time.Second)
	}
	last_xpath = ""
	log.Println("Не найден " + xpath)
	return false
}

// поиск на странице xpath из списка xpaths
func find_web_element_by_xpaths(xpaths []string) {
	if err != nil {
		return
	}

	for i := 0; i < max_wait_sec; i++ {
		for _, last_xpath = range xpaths {
			last_web_element, err = web_driver.FindElement(selenium.ByXPATH, last_xpath)
			if err == nil {
				return
			}
		}
		time.Sleep(time.Second)
	}
	last_xpath = ""
	err = fmt.Errorf("не найден ни один элемент xpath из %d", xpaths)
}

//поиск xpath на странице, ожидая max_wait_sec секунд
func find_web_element_array(xpath string) {
	//if err != nil {
	//	return
	//}

	for i := 0; i < max_wait_sec; i++ {
		last_web_elements, err = web_driver.FindElements(selenium.ByXPATH, xpath)
		if err == nil && len(last_web_elements) > 0 {
			return
		}
		time.Sleep(time.Second)
	}
}

//поиск на странице xpath и нажатие  на элемент
//если xpath не задан "", то без поиска, сразу нажатие
//если второй символ xpath не "/" то хpath, берется из xpaths[xpath]
func press_button(xpath string) {
	if err != nil {
		return
	}

	if xpath != "" {
		if xpath[1:2] != "/" {
			xpath = xpaths[xpath]
		}

		find_web_element(xpath)
	}

	if err == nil {
		err = last_web_element.Click()
	}
}

//подготовка к логированию на экран и в файл
func prepare_log(name_log_file string) *os.File {
	f, err := os.OpenFile(name_log_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	//defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	log.Println("Info Начало работы.")
	return f
}
