# Новые функции shell2telegram

## 1. Загрузка и скачивание файлов

### Загрузка файлов от пользователя
Отправьте любой файл в бот, и путь к нему будет доступен в переменной окружения `$S2T_FILE_PATH`.

**Пример:**
```bash
shell2telegram /process '#!/bin/bash
if [ -n "$S2T_FILE_PATH" ]; then
  echo "Processing file: $S2T_FILE_PATH"
  wc -l "$S2T_FILE_PATH"
else
  echo "Please upload a file"
fi'
```

### Скачивание файлов пользователю
Есть два способа отправить файл пользователю:

**Способ 1: Модификатор `:file`**
```bash
shell2telegram /report:file 'generate_report.sh > /tmp/report.txt && echo /tmp/report.txt'
```

**Способ 2: Префикс `FILE:` в выводе**
```bash
shell2telegram /backup 'backup.sh && echo "FILE:/path/to/backup.tar.gz"'
```

### Автоочистка
Загруженные файлы автоматически удаляются через 1 час.

---

## 2. История команд

### Просмотр истории
```bash
/history              # Показать последние 10 команд
/history export       # Экспорт всей истории в файл
```

### Как это работает
- Автоматически сохраняются последние 50 команд каждого пользователя
- Показывается: время, команда, аргументы, успешность выполнения (✓/✗)
- История сохраняется в JSON вместе с данными пользователей

---

## 3. Логирование в базу данных SQLite

### Включение логирования
```bash
shell2telegram --enable-db-logging [--db-path=/path/to/db.sqlite] /cmd 'command'
```

### Команды для работы с логами (только для root)

**Просмотр последних логов:**
```bash
/shell2telegram logs           # Последние 20 записей
/shell2telegram logs 50        # Последние 50 записей (макс 100)
```

**Поиск в логах:**
```bash
/shell2telegram search_logs backup   # Поиск по ключевому слову
```

**Статистика:**
```bash
/shell2telegram db_stats      # Общая статистика использования
```

### Что записывается в БД
- Timestamp (время выполнения)
- User ID и username
- Команда и аргументы
- Preview первых 500 символов вывода
- Время выполнения (мс)
- Успешность выполнения

---

## 4. Inline-кнопки (будущая функция)

⚠️ **Эта функция отложена** и требует обновления библиотеки telegram-bot-api до v5+.

В будущих версиях будет доступен синтаксис:
```bash
shell2telegram /menu:buttons="Start:start,Stop:stop,Status:status" 'handle_command.sh'
```

---

## Примеры использования

### Пример 1: Обработка загруженных файлов
```bash
shell2telegram /:plain_text:desc="Process uploaded files" '#!/bin/bash
if [ -n "$S2T_FILE_PATH" ]; then
  file_type=$(file -b "$S2T_FILE_PATH")
  size=$(du -h "$S2T_FILE_PATH" | cut -f1)
  lines=$(wc -l < "$S2T_FILE_PATH" 2>/dev/null || echo "N/A")
  echo "File type: $file_type"
  echo "Size: $size"
  echo "Lines: $lines"
else
  echo "Please upload a file to process"
fi'
```

### Пример 2: Генерация отчета с экспортом
```bash
shell2telegram /report:file:desc="Generate system report" '#!/bin/bash
report="/tmp/report_$(date +%Y%m%d_%H%M%S).txt"
echo "System Report" > "$report"
echo "=============" >> "$report"
echo "" >> "$report"
uptime >> "$report"
df -h >> "$report"
echo "$report"
```

### Пример 3: С включенным логированием и историей
```bash
shell2telegram \
  --enable-db-logging \
  --persistent-users \
  /status 'uptime' \
  /disk 'df -h' \
  /memory 'free -h'
```

Пользователи могут:
- Выполнять команды `/status`, `/disk`, `/memory`
- Просматривать свою историю через `/history`
- Root-пользователи могут просматривать все логи через `/shell2telegram logs`

---

## Переменные окружения

При выполнении команд доступны следующие переменные:

- `$S2T_LOGIN` - Telegram username пользователя
- `$S2T_USERID` - Telegram User ID
- `$S2T_USERNAME` - Имя пользователя (First + Last name)
- `$S2T_CHATID` - ID чата
- `$S2T_FILE_PATH` - Путь к загруженному файлу (если есть)

---

## Совместимость

- ✅ Загрузка/скачивание файлов - работает с telegram-bot-api v2
- ✅ История команд - работает с telegram-bot-api v2
- ✅ Логирование в БД - работает с telegram-bot-api v2
- ⚠️ Inline-кнопки - требует telegram-bot-api v5+ (будущая версия)
