# Инструкция по использованию shell2telegram с корпоративным прокси-сервером

## Изменения в программе

Добавлены следующие параметры командной строки для работы через прокси-сервер:

- `-proxy-server` - адрес прокси-сервера (можно указать как `host:port` или полный URL `http://host:port`)
- `-proxy-user` - имя пользователя для авторизации на прокси
- `-proxy-password` - пароль для авторизации на прокси

## Примеры использования

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

## Рекомендации для Windows Server

1. **Безопасность**: Рекомендуется хранить пароль прокси в переменной окружения или использовать bat-файл с ограниченными правами доступа:

```cmd
@echo off
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=your_password
set TB_TOKEN=your_bot_token

shell2telegram.exe -proxy-server %PROXY_SERVER% -proxy-user %PROXY_USER% -proxy-password %PROXY_PASSWORD% /date "date /t" /time "time /t"
```

2. **Запуск как служба Windows**: Можно использовать NSSM (Non-Sucking Service Manager) для запуска бота как службы Windows.

3. **Логирование**: Используйте параметр `-log` для записи логов:

```cmd
shell2telegram.exe -proxy-server proxy.company.com:8080 -proxy-user username -proxy-password password -log bot.log /date "date /t"
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
