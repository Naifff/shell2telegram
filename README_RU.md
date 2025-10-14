# 🤖 shell2telegram - Создайте Telegram бота из командной строки

[![Go build status](https://github.com/msoap/shell2telegram/actions/workflows/go.yml/badge.svg)](https://github.com/msoap/shell2telegram/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/msoap/shell2telegram/badge.svg?branch=master)](https://coveralls.io/github/msoap/shell2telegram?branch=master)

**shell2telegram** — это мощная утилита для создания Telegram ботов, которые выполняют shell-команды. Идеально подходит для администрирования серверов, автоматизации задач и удаленного управления системой через Telegram.

## 📋 Содержание

- [Установка](#установка)
- [Быстрый старт](#быстрый-старт)
- [Параметры командной строки](#параметры-командной-строки)
- [Работа через корпоративный прокси](#работа-через-корпоративный-прокси)
- [Модификаторы команд](#модификаторы-команд)
- [Системные переменные окружения](#системные-переменные-окружения)
- [Встроенные команды бота](#встроенные-команды-бота)
- [Примеры использования](#примеры-использования)
- [Продвинутые сценарии](#продвинутые-сценарии)
- [Безопасность](#безопасность)
- [Советы и лучшие практики](#советы-и-лучшие-практики)

## 🚀 Установка

### Windows (готовый бинарник)
Скачайте `shell2telegram.exe` из релизов или используйте собранный бинарник в этом репозитории.

### MacOS
```bash
brew tap msoap/tools
brew install shell2telegram
```

### Linux
```bash
go install github.com/msoap/shell2telegram@latest
```

### Docker
```dockerfile
FROM msoap/shell2telegram
ENV TB_TOKEN=your_token_here
CMD ["/date", "date"]
```

## ⚡ Быстрый старт

1. **Создайте бота через [@BotFather](https://t.me/BotFather)** и получите токен

2. **Запустите простого бота:**

**Linux/Mac:**
```bash
export TB_TOKEN="ваш_токен_от_BotFather"
shell2telegram /date 'date' /hello 'echo "Привет из shell2telegram!"'
```

**Windows:**
```cmd
set TB_TOKEN=ваш_токен_от_BotFather
shell2telegram.exe /date "date /t" /time "time /t"
```

3. **Найдите бота в Telegram** и отправьте `/help` — увидите список доступных команд

## 📝 Параметры командной строки

### Основные параметры

| Параметр | Описание | Пример |
|----------|----------|--------|
| `-tb-token` | Токен бота (или используйте переменную TB_TOKEN) | `-tb-token=123:ABC` |
| `-allow-users` | Разрешенные пользователи (через запятую) | `-allow-users=user1,user2` |
| `-root-users` | Root-пользователи для управления ботом | `-root-users=admin` |
| `-allow-all` | Разрешить всем пользователям (⚠️ ОПАСНО!) | `-allow-all` |
| `-log-commands` | Логировать все команды | `-log-commands` |
| `-log` | Файл для записи логов | `-log=bot.log` |
| `-add-exit` | Добавить команду `/shell2telegram exit` | `-add-exit` |
| `-persistent-users` | Сохранять пользователей в файл | `-persistent-users` |
| `-users-db` | Путь к файлу БД пользователей | `-users-db=users.json` |
| `-cache` | Кэшировать результаты команд (в секундах) | `-cache=60` |
| `-sh-timeout` | Таймаут выполнения команды (в секундах) | `-sh-timeout=30` |
| `-shell` | Указать shell для выполнения команд | `-shell=bash` |
| `-one-thread` | Выполнять команды последовательно | `-one-thread` |
| `-public` | Публичный бот (без авторизации) | `-public` |
| `-description` | Описание бота | `-description="Мой бот"` |

### 🌐 Параметры для работы через прокси (НОВОЕ!)

| Параметр | Описание | Пример |
|----------|----------|--------|
| `-proxy-server` | Адрес прокси-сервера | `-proxy-server=proxy.company.com:8080` |
| `-proxy-user` | Имя пользователя для прокси | `-proxy-user=username` |
| `-proxy-password` | Пароль для прокси | `-proxy-password=password` |

### Параметры для webhook

| Параметр | Описание |
|----------|----------|
| `-bind-addr` | Адрес для прослушивания webhook запросов |
| `-webhook` | URL для регистрации webhook |

## 🔐 Работа через корпоративный прокси

Если ваш сервер работает за корпоративным прокси (типично для Windows Server в корпоративной сети):

**Без авторизации:**
```cmd
shell2telegram.exe -proxy-server=proxy.company.com:8080 -tb-token=YOUR_TOKEN /ping "ping 8.8.8.8 -n 4"
```

**С авторизацией:**
```cmd
shell2telegram.exe -proxy-server=proxy.company.com:8080 -proxy-user=username -proxy-password=password -tb-token=YOUR_TOKEN /date "date /t"
```

**Безопасный способ (через bat-файл):**
```cmd
@echo off
REM start_bot.bat
set TB_TOKEN=ваш_токен
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=password

shell2telegram.exe ^
    -proxy-server=%PROXY_SERVER% ^
    -proxy-user=%PROXY_USER% ^
    -proxy-password=%PROXY_PASSWORD% ^
    -log-commands ^
    -log=bot.log ^
    -persistent-users ^
    /status "systeminfo | findstr /C:\"System Up Time\"" ^
    /disk "wmic logicaldisk get caption,freespace,size /format:list" ^
    /services "net start"
```

## 🎯 Модификаторы команд

### `:desc` — Описание команды
```bash
shell2telegram /start:desc="Начало работы" 'echo "Добро пожаловать!"'
```

### `:vars` — Переменные вместо STDIN
```bash
# Вместо отправки текста в STDIN, создаются переменные окружения
shell2telegram /search:vars=TERM 'grep -r "$TERM" /var/log'
# Использование: /search error
```

### `:md` — Форматирование Markdown
```bash
shell2telegram /info:md 'echo "*Жирный* _курсив_ \`код\`"'
```

### Комбинирование модификаторов
```bash
shell2telegram /cal:desc="Календарь":md 'echo "\`\`\`$(cal)\`\`\`"'
```

## 🔧 Системные переменные окружения

В каждой команде доступны специальные переменные:

- `$S2T_LOGIN` — @username пользователя Telegram
- `$S2T_USERID` — ID пользователя Telegram
- `$S2T_USERNAME` — Имя и фамилия пользователя
- `$S2T_CHATID` — ID чата

**Пример использования:**
```bash
shell2telegram /whoami 'echo "Вы: $S2T_USERNAME (@$S2T_LOGIN), ID: $S2T_USERID"'
```

## 📱 Встроенные команды бота

### Для всех пользователей
- `/help` — список доступных команд
- `/auth <CODE>` — авторизация по коду
- `/authroot <CODE>` — авторизация как root

### Для root-пользователей
- `/shell2telegram stat` — статистика пользователей
- `/shell2telegram search <query>` — поиск пользователей
- `/shell2telegram ban <user_id|@username>` — заблокировать пользователя
- `/shell2telegram exit` — завершить работу бота (требует `-add-exit`)
- `/shell2telegram desc <text>` — изменить описание бота
- `/shell2telegram rm </command>` — удалить команду
- `/shell2telegram broadcast_to_root <msg>` — сообщение всем root-пользователям
- `/shell2telegram message_to_user <user> <msg>` — личное сообщение пользователю
- `/shell2telegram version` — версия бота

## 💡 Примеры использования

### 1. 🖥️ Мониторинг Windows Server

```cmd
shell2telegram.exe ^
    -proxy-server=proxy:8080 ^
    -proxy-user=user ^
    -proxy-password=pass ^
    -root-users=admin ^
    -log-commands ^
    -persistent-users ^
    /cpu:desc="Загрузка CPU" "wmic cpu get loadpercentage /value" ^
    /memory:desc="Память" "systeminfo | findstr /C:\"Available Physical Memory\" /C:\"Total Physical Memory\"" ^
    /disk:desc="Диски" "wmic logicaldisk get caption,freespace,size /format:table" ^
    /services:desc="Службы" "net start" ^
    /uptime:desc="Время работы" "systeminfo | findstr /C:\"System Boot Time\"" ^
    /processes:desc="Процессы" "tasklist | sort /R /+65" ^
    /restart:desc="⚠️ Перезагрузка" "shutdown /r /t 60 /c \"Перезагрузка через Telegram\""
```

### 2. 📊 Мониторинг Linux сервера

```bash
export TB_TOKEN="ваш_токен"
shell2telegram \
    -root-users=admin \
    -log-commands \
    -persistent-users \
    -cache=30 \
    /status:desc="Статус системы" 'uptime' \
    /cpu:desc="CPU загрузка" 'top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk "{print 100 - \$1\"%\"}"' \
    /memory:desc="Память" 'free -h' \
    /disk:desc="Диски" 'df -h' \
    /top:desc="TOP процессы" 'ps aux --sort=-%cpu | head -20' \
    /network:desc="Сеть" 'ss -tulpn' \
    /users:desc="Пользователи" 'who'
```

### 3. 🔔 Система уведомлений и таймеры

```bash
shell2telegram \
    /remind:vars=MINUTES,TEXT:desc="Напоминание" 'sleep $(($MINUTES * 60)); echo "⏰ Напоминание: $TEXT"' \
    /timer:vars=SECONDS:desc="Таймер" 'sleep $SECONDS; echo "✅ Таймер завершен!"' \
    /alarm:vars=TIME,MSG:desc="Будильник" 'sleep $TIME && echo "🔔 $MSG"'
```

**Использование:**
- `/remind 5 Проверить почту` — напомнит через 5 минут
- `/timer 60` — таймер на 60 секунд

### 4. 📁 Управление файлами и резервными копиями

**Windows:**
```cmd
shell2telegram.exe ^
    /backup:desc="Создать бэкап" "robocopy C:\Important D:\Backup\Important /MIR /LOG:backup.log && type backup.log | findstr /C:\"Files\" /C:\"Dirs\"" ^
    /listbackups:desc="Список бэкапов" "dir D:\Backup /B" ^
    /checkspace:desc="Свободное место" "dir C:\ | findstr \"bytes free\""
```

**Linux:**
```bash
shell2telegram \
    /backup:desc="Бэкап базы" 'tar -czf /backup/db_$(date +%Y%m%d).tar.gz /var/lib/mysql && ls -lh /backup/*.tar.gz | tail -5' \
    /listbackups:desc="Список бэкапов" 'ls -lh /backup/*.tar.gz' \
    /cleanold:desc="Удалить старые" 'find /backup -name "*.tar.gz" -mtime +7 -delete && echo "Старые бэкапы удалены"'
```

### 5. 🌐 Сетевая диагностика

```bash
shell2telegram \
    /ping:vars=HOST:desc="Пинг хоста" 'ping -c 4 $HOST' \
    /trace:vars=HOST:desc="Трассировка" 'traceroute $HOST' \
    /ports:desc="Открытые порты" 'netstat -tulpn' \
    /checksite:vars=URL:desc="Проверка сайта" 'curl -I $URL | head -1' \
    /dns:vars=DOMAIN:desc="DNS lookup" 'nslookup $DOMAIN' \
    /myip:desc="Мой внешний IP" 'curl -s ifconfig.me'
```

### 6. 🗄️ Работа с базами данных

**MySQL:**
```bash
shell2telegram \
    /dbstatus:desc="Статус MySQL" 'mysql -e "SHOW STATUS LIKE \"Threads_connected\";"' \
    /dbsize:desc="Размер БД" 'mysql -e "SELECT table_schema, ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS \"Size (MB)\" FROM information_schema.tables GROUP BY table_schema;"' \
    /slowqueries:desc="Медленные запросы" 'tail -50 /var/log/mysql/slow.log'
```

**PostgreSQL:**
```bash
shell2telegram \
    /pgstatus:desc="PostgreSQL статус" 'systemctl status postgresql' \
    /pgactivity:desc="Активность" 'psql -c "SELECT pid, usename, application_name, state FROM pg_stat_activity;"'
```

### 7. 📝 Просмотр и анализ логов

**Windows:**
```cmd
shell2telegram.exe ^
    /logs:vars=LINES:desc="Последние логи" "powershell Get-Content bot.log -Tail $env:LINES" ^
    /errors:desc="Ошибки" "findstr /C:\"error\" /C:\"ERROR\" bot.log | findstr /C:\"$(Get-Date -Format \"yyyy-MM-dd\")\"" ^
    /eventlog:desc="System Event Log" "powershell Get-EventLog -LogName System -Newest 10 | Format-Table -AutoSize"
```

**Linux:**
```bash
shell2telegram \
    /logs:vars=LINES:desc="Последние логи" 'tail -n ${LINES:-50} /var/log/syslog' \
    /errors:desc="Ошибки" 'grep -i error /var/log/syslog | tail -20' \
    /apache:desc="Apache логи" 'tail -50 /var/log/apache2/error.log' \
    /nginx:desc="Nginx логи" 'tail -50 /var/log/nginx/error.log'
```

### 8. 🔄 Управление службами

**Windows:**
```cmd
shell2telegram.exe ^
    /services:desc="Все службы" "sc query type= service state= all | findstr SERVICE_NAME" ^
    /restart:vars=SERVICE:desc="Перезапуск службы" "net stop $env:SERVICE && net start $env:SERVICE" ^
    /status:vars=SERVICE:desc="Статус службы" "sc query $env:SERVICE"
```

**Linux:**
```bash
shell2telegram \
    /services:desc="Все службы" 'systemctl list-units --type=service --state=running' \
    /restart:vars=SERVICE:desc="Перезапуск" 'sudo systemctl restart $SERVICE && systemctl status $SERVICE' \
    /stop:vars=SERVICE:desc="Остановить" 'sudo systemctl stop $SERVICE' \
    /start:vars=SERVICE:desc="Запустить" 'sudo systemctl start $SERVICE'
```

### 9. 🔍 Мониторинг приложений

```bash
shell2telegram \
    /dockerps:desc="Docker контейнеры" 'docker ps -a' \
    /dockerstats:desc="Docker статистика" 'docker stats --no-stream' \
    /dockerlogs:vars=CONTAINER:desc="Логи контейнера" 'docker logs --tail 50 $CONTAINER' \
    /gitlog:desc="Git история" 'cd /var/www/app && git log --oneline -10' \
    /gitstatus:desc="Git статус" 'cd /var/www/app && git status' \
    /gitpull:desc="Git pull" 'cd /var/www/app && git pull'
```

### 10. 📊 Статистика и отчеты

```bash
shell2telegram \
    /report:desc="Дневной отчет":md 'echo "**📊 Отчет за $(date +%d.%m.%Y)**\n\n**CPU:** $(uptime | awk -F\"load average:\" \"{print \$2}\")\n**Memory:** $(free -h | awk \"NR==2{print \$3\"/\"\$2}\")\n**Disk:** $(df -h / | awk \"NR==2{print \$5}\")"' \
    /visits:desc="Посещения сайта" 'tail -1000 /var/log/nginx/access.log | wc -l' \
    /traffic:desc="Трафик" 'vnstat -d' \
    /bandwidth:desc="Пропускная способность" 'speedtest-cli --simple'
```

### 11. 🎨 Интерактивные команды

```bash
# Получение любого текста от пользователя
shell2telegram \
    /:plain_text:desc="Сортировка текста" 'sort' \
    /reverse:desc="Реверс текста" 'rev' \
    /upper:desc="В ВЕРХНИЙ регистр" 'tr a-z A-Z' \
    /count:desc="Подсчет слов" 'wc -w'
```

**Использование:**
1. Отправьте команду `/upper`
2. Затем отправьте любой текст (без команды)
3. Получите результат в ВЕРХНЕМ РЕГИСТРЕ

### 12. 🔐 Безопасность и аудит

```bash
shell2telegram \
    /lastlogins:desc="Последние входы" 'last -10' \
    /failedlogins:desc="Неудачные входы" 'grep "Failed password" /var/log/auth.log | tail -20' \
    /openfiles:desc="Открытые файлы" 'lsof | wc -l' \
    /connections:desc="Сетевые соединения" 'netstat -an | grep ESTABLISHED | wc -l' \
    /firewall:desc="Правила firewall" 'iptables -L -n'
```

## 🚀 Продвинутые сценарии

### CI/CD через Telegram

```bash
shell2telegram \
    -root-users=devops_team \
    /deploy:vars=BRANCH:desc="🚀 Деплой" 'cd /var/www/app && git fetch && git checkout $BRANCH && git pull && npm install && npm run build && systemctl restart app && echo "✅ Деплой ветки $BRANCH завершен!"' \
    /rollback:desc="⏪ Откат" 'cd /var/www/app && git reset --hard HEAD~1 && systemctl restart app && echo "✅ Откат выполнен"' \
    /tests:desc="🧪 Тесты" 'cd /var/www/app && npm test' \
    /build:desc="🔨 Сборка" 'cd /var/www/app && npm run build'
```

### Мониторинг с уведомлениями

```bash
shell2telegram \
    /checkdisk:desc="Проверка дисков" 'USAGE=$(df -h / | awk "NR==2{print \$5}" | sed "s/%//"); if [ $USAGE -gt 80 ]; then echo "⚠️ ВНИМАНИЕ! Диск заполнен на $USAGE%"; else echo "✅ Диск в норме: $USAGE%"; fi' \
    /checkmemory:desc="Проверка памяти" 'FREE=$(free | awk "NR==2{printf \"%.0f\", \$4/\$2*100}"); if [ $FREE -lt 20 ]; then echo "⚠️ Мало памяти! Свободно: $FREE%"; else echo "✅ Память в норме: $FREE%"; fi'
```

### Удаленное управление с подтверждением

```bash
shell2telegram \
    -root-users=admin \
    /reboot:desc="🔄 Перезагрузка (используй /reboot confirm)" 'if [ "$1" = "confirm" ]; then shutdown -r +1 "Перезагрузка через 1 минуту"; echo "✅ Перезагрузка запланирована"; else echo "⚠️ Используйте: /reboot confirm"; fi' \
    /cancelreboot:desc="❌ Отмена перезагрузки" 'shutdown -c; echo "✅ Перезагрузка отменена"'
```

## 🔒 Безопасность

### Рекомендации по безопасности:

1. **Всегда используйте авторизацию:**
   ```bash
   -root-users=admin -allow-users=user1,user2
   ```

2. **Для Windows Server храните пароли в защищенном bat-файле:**
   ```cmd
   REM Установите права доступа только для Администраторов
   icacls start_bot.bat /inheritance:r /grant:r Administrators:F
   ```

3. **Используйте таймауты для длительных операций:**
   ```bash
   -sh-timeout=30
   ```

4. **Логируйте все команды:**
   ```bash
   -log-commands -log=bot_audit.log
   ```

5. **Ограничьте опасные команды только для root:**
   Проверяйте, что опасные команды (shutdown, rm -rf, etc.) выполняют только root-пользователи

6. **Регулярно проверяйте список пользователей:**
   ```bash
   /shell2telegram stat
   ```

## 💎 Советы и лучшие практики

### 1. Кэширование для частых запросов
```bash
# Кэшировать результат на 60 секунд
shell2telegram -cache=60 /status 'uptime'
```

### 2. Последовательное выполнение критичных команд
```bash
# Если команды не должны выполняться параллельно
shell2telegram -one-thread /backup 'backup_script.sh'
```

### 3. Использование webhook для высокой нагрузки
```bash
shell2telegram -bind-addr=0.0.0.0:8443 -webhook=https://yourdomain.com:8443/bot /cmd 'command'
```

### 4. Автоматический перезапуск (Linux)

**Systemd service** (`/etc/systemd/system/telegram-bot.service`):
```ini
[Unit]
Description=Telegram Shell Bot
After=network.target

[Service]
Type=simple
User=botuser
WorkingDirectory=/opt/bot
Environment="TB_TOKEN=your_token"
ExecStart=/usr/local/bin/shell2telegram -log=/var/log/telegram-bot.log /date 'date'
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Запуск:
```bash
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

### 5. Автоматический запуск (Windows)

Используйте **Task Scheduler** или **NSSM** (Non-Sucking Service Manager):
```cmd
nssm install TelegramBot "C:\bot\shell2telegram.exe" -log=bot.log /date "date /t"
nssm start TelegramBot
```

### 6. Мониторинг работы бота
```bash
# Добавьте команду проверки живости
shell2telegram /healthcheck 'echo "Bot is alive! $(date)"'
```

### 7. Форматирование вывода для читаемости
```bash
shell2telegram \
    /status:md 'echo "\`\`\`"; uptime; free -h; df -h; echo "\`\`\`"'
```

### 8. Работа с конфиденциальными данными
```bash
# НЕ передавайте пароли в параметрах! Используйте переменные:
export PROXY_PASSWORD="secret"
export TB_TOKEN="bot_token"
shell2telegram -proxy-password="$PROXY_PASSWORD" /cmd 'command'
```

## 🐛 Troubleshooting

### Бот не отвечает
1. Проверьте токен: убедитесь, что TB_TOKEN корректный
2. Проверьте прокси настройки (для корпоративной сети)
3. Проверьте логи: `tail -f bot.log`

### Команды выполняются медленно
1. Увеличьте `-timeout`
2. Используйте `-cache` для часто запрашиваемых данных
3. Оптимизируйте shell команды

### Ошибки авторизации прокси
```bash
# Убедитесь, что логин и пароль указаны правильно
# Проверьте, нужен ли домен: -proxy-user=DOMAIN\username
```

## 📚 Дополнительные ресурсы

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Документация BotFather](https://core.telegram.org/bots#botfather)
- [GitHub репозиторий](https://github.com/msoap/shell2telegram)
- [Telegram канал о shell2telegram](https://t.me/shell2telegram)

## 📄 Лицензия

MIT License - используйте свободно в коммерческих и личных проектах!

---

**Создано с ❤️ для администраторов и DevOps инженеров**

*Если вам понравился этот проект, поставьте ⭐ на GitHub!*
