# gosuslugi_request_sim.exe
По ФЗ-533 до 01.12.2021 нужно зарегистрировать симкарты на госуслугах.
ПО для автоматизации подтверждения сразу всех номеров телефонов от Мегафон и Теле2:<br>
Открытый исходный код на языке go:<br>
https://github.com/san035/gosuslugi_active_sim<br>

<b>Что делает программа gosuslugi_request_sim.exe:</b><br>
Открывает в браузере Chrome сайт gosuslugi.ru<br>
Подставляет логин/пароль из config.ini или переменных окружения gosuslugi_login/gosuslugi_password<br>
Находит все входящие сообщения с темой "Запрос на активацию корпоративной сим-карты"<br>
Поочереди открывает их и подтверждает запрос.<br>

<b>Сохранение логин/пароль в переменных окружения windows:</b><br>
setx gosuslugi_login 00011122233<br>
setx gosuslugi_password your_password<br>

В файле sim.log останется отчет какие симки были подтверждены.<br>

Для работы программы нужен chromedriver.exe соответсвующий установленоой версии браузера chrome.exe, ссылка для обновления chromedriver.exe:<br>
https://chromedriver.chromium.org/downloads
