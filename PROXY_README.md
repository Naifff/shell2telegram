# Инструкция по использованию shell2telegram с корпоративным прокси-сервером

## Изменения в программе

Добавлена поддержка работы через прокси-сервер **двумя способами**:

### Способ 1: Параметры командной строки

- `-proxy-server` - адрес прокси-сервера (можно указать как `host:port` или полный URL `http://host:port`)
- `-proxy-user` - имя пользователя для авторизации на прокси
- `-proxy-password` - пароль для авторизации на прокси

### Способ 2: Переменные окружения (НОВОЕ!)

Более безопасный способ, особенно для Windows Server:

**Кастомные переменные:**
- `PROXY_SERVER` - адрес прокси-сервера
- `PROXY_USER` - имя пользователя
- `PROXY_PASSWORD` - пароль

**Стандартные переменные (совместимость с другими приложениями):**
- `HTTP_PROXY` или `http_proxy` - адрес прокси (может содержать учетные данные)
- `HTTPS_PROXY` или `https_proxy` - адрес прокси для HTTPS

**Приоритет настроек:**
1. Параметры командной строки (высший приоритет)
2. Кастомные переменные окружения (PROXY_SERVER, PROXY_USER, PROXY_PASSWORD)
3. Стандартные переменные окружения (HTTP_PROXY, HTTPS_PROXY)

## Примеры использования

### Способ А: Параметры командной строки

### 1. Простой прокси без авторизации

```cmd
shell2telegram.exe -proxy-server proxy.company.com:8080 -tb-token YOUR_BOT_TOKEN /date "date /t"
```

### 2. Прокси с авторизацией (рекомендуется для Windows Server)

```cmd
shell2telegram.exe -proxy-server proxy.company.com:8080 -proxy-user username -proxy-password password -tb-token YOUR_BOT_TOKEN /date "date /t" /time "time /t"
```

### 3. С указанием полного URL прокси

```cmd
shell2telegram.exe -proxy-server http://proxy.company.com:8080 -proxy-user username -proxy-password password -tb-token YOUR_BOT_TOKEN /date "date /t"
```

### 4. Использование переменной окружения для токена

```cmd
set TB_TOKEN=YOUR_BOT_TOKEN
shell2telegram.exe -proxy-server proxy.company.com:8080 -proxy-user username -proxy-password password /date "date /t" /ipconfig "ipconfig"
```

### Способ Б: Переменные окружения (РЕКОМЕНДУЕТСЯ!)

Это более безопасный способ, так как пароли не видны в командной строке.

### 1. Использование кастомных переменных (Windows)

```cmd
REM Установка переменных окружения
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=your_password
set TB_TOKEN=your_bot_token

REM Запуск бота (без указания прокси в параметрах!)
shell2telegram.exe /date "date /t" /time "time /t" /ipconfig "ipconfig"
```

### 2. Использование стандартных переменных HTTP_PROXY (Windows)

```cmd
REM Вариант 1: С учетными данными в URL
set HTTP_PROXY=http://username:password@proxy.company.com:8080
set TB_TOKEN=your_bot_token
shell2telegram.exe /date "date /t"

REM Вариант 2: Без учетных данных в URL + отдельные переменные
set HTTP_PROXY=http://proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=password
set TB_TOKEN=your_bot_token
shell2telegram.exe /date "date /t"
```

### 3. Постоянные переменные окружения (Windows)

Для установки переменных на уровне системы:

```cmd
REM Запустите от имени Администратора
setx PROXY_SERVER "proxy.company.com:8080" /M
setx PROXY_USER "username" /M
setx PROXY_PASSWORD "your_password" /M
setx TB_TOKEN "your_bot_token" /M

REM После перезапуска командной строки:
shell2telegram.exe /date "date /t"
```

### 4. Использование в Linux/Mac

```bash
# Экспорт переменных окружения
export PROXY_SERVER=proxy.company.com:8080
export PROXY_USER=username
export PROXY_PASSWORD=password
export TB_TOKEN=your_bot_token

# Или использование стандартных переменных
export HTTP_PROXY=http://username:password@proxy.company.com:8080
export TB_TOKEN=your_bot_token

# Запуск бота
./shell2telegram /date 'date' /uptime 'uptime'
```

### 5. Комбинирование способов

```cmd
REM Можно установить прокси через переменные окружения,
REM но переопределить через параметры командной строки:

set HTTP_PROXY=http://old-proxy:8080
set TB_TOKEN=your_token

REM Этот запуск будет использовать new-proxy вместо old-proxy:
shell2telegram.exe -proxy-server new-proxy.company.com:3128 /date "date /t"
```

## Рекомендации для Windows Server

### 1. **Безопасность**: Использование bat-файла с переменными окружения

Создайте файл `start_bot.bat` с ограниченными правами доступа:

```cmd
@echo off
REM start_bot.bat - Запуск Telegram бота с прокси
REM Установите права доступа: icacls start_bot.bat /inheritance:r /grant:r Administrators:F

REM Настройка прокси через переменные окружения
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=your_password
set TB_TOKEN=your_bot_token

REM Запуск бота (параметры прокси берутся из переменных окружения автоматически!)
shell2telegram.exe ^
    -log-commands ^
    -log=bot.log ^
    -persistent-users ^
    -root-users=admin ^
    /status:desc="Статус системы" "systeminfo | findstr /C:\"System Up Time\"" ^
    /disk:desc="Диски" "wmic logicaldisk get caption,freespace,size /format:table" ^
    /memory:desc="Память" "systeminfo | findstr /C:\"Available Physical Memory\"" ^
    /services:desc="Службы" "net start" ^
    /date "date /t" ^
    /time "time /t"

pause
```

**Установка прав доступа (запустите от Администратора):**
```cmd
icacls start_bot.bat /inheritance:r /grant:r Administrators:F
```

### 2. **Запуск как служба Windows**

Можно использовать **NSSM** (Non-Sucking Service Manager):

```cmd
REM Скачайте NSSM с https://nssm.cc/

REM Установите переменные окружения на уровне системы
setx PROXY_SERVER "proxy.company.com:8080" /M
setx PROXY_USER "username" /M
setx PROXY_PASSWORD "password" /M
setx TB_TOKEN "your_token" /M

REM Установите службу
nssm install TelegramBot "C:\bot\shell2telegram.exe" -log-commands -log C:\bot\bot.log /date "date /t" /time "time /t"

REM Запустите службу
nssm start TelegramBot

REM Проверка статуса
nssm status TelegramBot
```

### 3. **Логирование с переменными окружения**

```cmd
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=password
set TB_TOKEN=your_token

shell2telegram.exe -log bot.log -log-commands /date "date /t"
```

## Сборка бинарника

Бинарник для Windows 64-bit уже собран: `shell2telegram.exe`

Если нужно пересобрать:

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o shell2telegram.exe
```

Или через Docker:

```bash
docker run --rm -v "$PWD":/go/src/app -w /go/src/app golang:alpine sh -c "apk add --no-cache git && GOOS=windows GOARCH=amd64 go build -ldflags='-w -s' -o shell2telegram.exe"
```

## Проверка работы

После запуска в логах должно появиться сообщение:
```
Using proxy server: proxy.company.com:8080
Authorized on bot account: @your_bot_name
```

Если прокси не настроен (параметры не указаны), бот будет работать с прямым подключением к интернету (как раньше).

## Справочная таблица: Поддерживаемые переменные окружения

| Переменная | Описание | Пример значения | Приоритет |
|------------|----------|-----------------|-----------|
| `PROXY_SERVER` | Адрес прокси-сервера (кастомная переменная) | `proxy.company.com:8080` | 2 |
| `PROXY_USER` | Имя пользователя для прокси | `username` или `DOMAIN\username` | - |
| `PROXY_PASSWORD` | Пароль для прокси | `your_password` | - |
| `HTTP_PROXY` | Стандартная переменная для HTTP прокси (может содержать учетные данные) | `http://user:pass@proxy:8080` | 3 |
| `http_proxy` | То же, но в нижнем регистре (для Linux) | `http://proxy:8080` | 4 |
| `HTTPS_PROXY` | Стандартная переменная для HTTPS прокси | `http://proxy:8080` | 5 |
| `https_proxy` | То же, но в нижнем регистре (для Linux) | `http://proxy:8080` | 6 |
| `TB_TOKEN` | Токен Telegram бота | `123456789:ABCdefGHIjklMNOpqrsTUVwxyz` | - |

**Примечание:** Приоритет 1 имеют параметры командной строки (`-proxy-server`, `-proxy-user`, `-proxy-password`).

## Форматы адреса прокси-сервера

Поддерживаются следующие форматы:

```
# Простой формат (добавится http:// автоматически)
proxy.company.com:8080
192.168.1.100:3128

# С протоколом
http://proxy.company.com:8080
https://secure-proxy.company.com:8443

# С учетными данными в URL (для HTTP_PROXY)
http://username:password@proxy.company.com:8080
http://DOMAIN%5Cusername:password@proxy:8080  (для Windows домена: DOMAIN\username)

# Только хост (порт 80 по умолчанию)
proxy.company.com
```

## Устранение проблем (Troubleshooting)

### 1. Бот не подключается через прокси

**Проверьте настройки:**
```cmd
REM Windows - показать текущие переменные окружения
echo %PROXY_SERVER%
echo %HTTP_PROXY%
echo %PROXY_USER%
```

```bash
# Linux/Mac
echo $PROXY_SERVER
echo $HTTP_PROXY
echo $PROXY_USER
```

**Проверьте логи:**
```cmd
shell2telegram.exe -log bot.log -log-commands /test "echo test"
type bot.log
```

### 2. Ошибка аутентификации прокси

- Убедитесь, что логин и пароль указаны правильно
- Для Windows домена используйте формат: `DOMAIN\username` или `DOMAIN%5Cusername` (в URL)
- Проверьте, не истек ли срок действия пароля

### 3. Переменные окружения не работают

**Windows:**
```cmd
REM Проверьте, что переменные установлены в текущей сессии
set | findstr PROXY
set | findstr TB_TOKEN

REM Для постоянных переменных - перезапустите командную строку после setx
```

**Linux/Mac:**
```bash
# Убедитесь, что переменные экспортированы
env | grep -i proxy
env | grep TB_TOKEN
```

### 4. Конфликт настроек

Если прокси настроен несколькими способами, помните о приоритете:
1. **Параметры командной строки** (высший приоритет)
2. **PROXY_SERVER** + PROXY_USER + PROXY_PASSWORD
3. **HTTP_PROXY / http_proxy**
4. **HTTPS_PROXY / https_proxy** (низший приоритет)

## Дополнительные возможности

### Отключение прокси для тестирования

```cmd
REM Временно отключить прокси (переопределить через параметр)
set HTTP_PROXY=http://proxy:8080
shell2telegram.exe -proxy-server="" /test "echo test"
```

### Использование разных прокси для разных ботов

```cmd
REM Бот 1 - через корпоративный прокси
start /B shell2telegram.exe -proxy-server corporate.proxy:8080 -proxy-user user1 -proxy-password pass1 -tb-token TOKEN1 /cmd1 "command1"

REM Бот 2 - без прокси
start /B shell2telegram.exe -tb-token TOKEN2 /cmd2 "command2"
```

## Безопасность

⚠️ **ВАЖНО:**

1. **Никогда не храните пароли в открытом виде в коде или скриптах, которые доступны другим пользователям**
2. Используйте переменные окружения вместо параметров командной строки (параметры видны в `ps`/`tasklist`)
3. Ограничьте права доступа к bat-файлам с паролями:
   ```cmd
   icacls start_bot.bat /inheritance:r /grant:r Administrators:F
   ```
4. Используйте системные переменные окружения (setx /M) только на доверенных серверах
5. Регулярно меняйте пароли прокси
