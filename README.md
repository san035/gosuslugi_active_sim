# gosuslugi_request_sim.exe
По ФЗ-533 до 01.12.2021 нужно зарегистрировать симкарты на госуслугах.
ПО для автоматизации подтверждения сразу всех номеров телефонов от Мегафона:<br>
Открытый исходный код на языке go:<br>
https://github.com/san035/gosuslugi_active_sim<br>

Что делает программа gosuslugi_request_sim.exe:<br>
Открывает в браузере Chrome сайт gosuslugi.ru
Подставляет логин/пароль из файла config.ini
Находит все входящие сообщения с темой "Запрос на активацию корпоративной сим-карты"
Поочереди открывает их и подтверждает запрос.
В файле sim.log останется отчет какие симки были подтверждены.
