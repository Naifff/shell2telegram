# Новая команда `/:file` - Автоматическая обработка файлов

## Что это?

`/:file` - это специальная команда, аналогичная `/:plain_text`, но **только для файлов**.

### Отличия от `/:plain_text`:

| Команда | Когда срабатывает | Для чего |
|---------|------------------|----------|
| `/:plain_text` | При любом тексте без `/` | Обработка текстовых сообщений |
| `/:file` | При загрузке файла | **Только обработка файлов** |

## Преимущества:

✅ `/:plain_text` свободна для текстовых задач
✅ `/:file` автоматически обрабатывает все файлы
✅ Можно использовать обе команды одновременно
✅ Чистое разделение логики

---

## Пример использования

### Вариант 1: Только обработка файлов

```bash
export TB_TOKEN="your_bot_token"

./shell2telegram_darwin_arm64 \
  --allow-all \
  /:file:desc="Auto-process uploaded files" 'bash -c "
echo \"✅ Файл получен!\"
echo \"━━━━━━━━━━━━━━━━━━━━━\"
echo \"📁 Имя: \$(basename \"\$S2T_FILE_PATH\")\"
echo \"📂 Путь: \$S2T_FILE_PATH\"
echo \"📊 Размер: \$(du -h \"\$S2T_FILE_PATH\" | cut -f1)\"
echo \"📝 Тип: \$(file -b \"\$S2T_FILE_PATH\")\"
echo \"━━━━━━━━━━━━━━━━━━━━━\"
echo \"\"

# Обработка текстовых файлов
if [[ \"\$S2T_FILE_PATH\" == *.txt ]] || [[ \"\$S2T_FILE_PATH\" == *.md ]]; then
  echo \"📄 Содержимое:\"
  cat \"\$S2T_FILE_PATH\"
  echo \"\"
  echo \"📊 Строк: \$(wc -l < \"\$S2T_FILE_PATH\")\"
  echo \"📊 Слов: \$(wc -w < \"\$S2T_FILE_PATH\")\"
fi
"'
```

**Тест:**
- Отправь файл → Автоматически обработается
- Отправь текст → Ничего не произойдет (нет `/:plain_text`)

---

### Вариант 2: Файлы + Текст (два обработчика)

```bash
./shell2telegram_darwin_arm64 \
  --allow-all \
  /:plain_text:desc="Process text messages" 'bash -c "
echo \"💬 Получено текстовое сообщение:\"
echo \"\"
echo \"\$*\"
echo \"\"
echo \"От: \$S2T_USERNAME\"
echo \"User ID: \$S2T_USERID\"
"' \
  /:file:desc="Process uploaded files" 'bash -c "
echo \"📁 Получен файл: \$(basename \"\$S2T_FILE_PATH\")\"
echo \"📊 Размер: \$(du -h \"\$S2T_FILE_PATH\" | cut -f1)\"
echo \"\"

# CSV анализ
if [[ \"\$S2T_FILE_PATH\" == *.csv ]]; then
  echo \"📊 CSV файл - первые 5 строк:\"
  head -5 \"\$S2T_FILE_PATH\"
fi

# JSON валидация
if [[ \"\$S2T_FILE_PATH\" == *.json ]]; then
  echo \"🔍 JSON файл - проверка:\"
  if python3 -m json.tool \"\$S2T_FILE_PATH\" > /dev/null 2>&1; then
    echo \"✅ Валидный JSON\"
  else
    echo \"❌ Невалидный JSON\"
  fi
fi
"'
```

**Тест:**
- Отправь текст "Hello" → Обработает `/:plain_text`
- Отправь файл → Обработает `/:file`
- Чистое разделение логики! 🎯

---

## Продвинутые примеры

### Пример 1: Обработчик изображений

```bash
./shell2telegram_darwin_arm64 \
  --allow-all \
  /:file 'bash -c "
FILE=\"\$S2T_FILE_PATH\"

if [[ \"\$FILE\" == *.jpg ]] || [[ \"\$FILE\" == *.png ]] || [[ \"\$FILE\" == *.jpeg ]]; then
  echo \"🖼️ Изображение получено!\"
  echo \"\"

  # Используем imagemagick для анализа (если установлен)
  if command -v identify > /dev/null; then
    echo \"📐 Размеры:\"
    identify -format \"Width: %w px\nHeight: %h px\n\" \"\$FILE\"
  fi

  echo \"📊 Размер файла: \$(du -h \"\$FILE\" | cut -f1)\"
else
  echo \"⚠️ Это не изображение\"
  echo \"Тип: \$(file -b \"\$FILE\")\"
fi
"'
```

### Пример 2: Логирование файлов в БД

```bash
./shell2telegram_darwin_arm64 \
  --allow-all \
  --enable-db-logging \
  /:file 'bash -c "
echo \"📁 Файл сохранен: \$(basename \"\$S2T_FILE_PATH\")\"
echo \"\"
echo \"Используй /shell2telegram logs для просмотра истории\"
"'
```

### Пример 3: Конвертер файлов

```bash
./shell2telegram_darwin_arm64 \
  --allow-all \
  /:file:desc="Auto-convert files" 'bash -c "
FILE=\"\$S2T_FILE_PATH\"
BASENAME=\$(basename \"\$FILE\")

echo \"🔄 Конвертация файла: \$BASENAME\"

if [[ \"\$FILE\" == *.md ]]; then
  # Markdown → HTML
  OUTPUT=\"/tmp/\${BASENAME%.md}.html\"
  if command -v pandoc > /dev/null; then
    pandoc \"\$FILE\" -o \"\$OUTPUT\"
    echo \"✅ Конвертировано в HTML\"
    echo \"FILE:\$OUTPUT\"
  fi
elif [[ \"\$FILE\" == *.txt ]]; then
  # TXT → PDF
  echo \"📝 Текстовый файл получен\"
  echo \"Конвертация в PDF не реализована\"
else
  echo \"ℹ️ Тип файла: \$(file -b \"\$FILE\")\"
fi
"'
```

---

## Комбинация с другими командами

```bash
./shell2telegram_darwin_arm64 \
  --allow-all \
  --persistent-users \
  /start 'echo "👋 Привет! Отправь мне файл для обработки"' \
  /help 'echo "Доступные команды:\n/start - начало\n/status - статус\n\nПросто отправь файл для автоматической обработки!"' \
  /status 'echo "🟢 Бот работает\nВремя: $(date)"' \
  /:file 'bash -c "
echo \"📁 Файл: \$(basename \"\$S2T_FILE_PATH\")\"
echo \"👤 Отправил: \$S2T_USERNAME\"
echo \"📊 Размер: \$(du -h \"\$S2T_FILE_PATH\" | cut -f1)\"
"'
```

---

## Переменные окружения

При использовании `/:file` доступны все стандартные переменные:

| Переменная | Описание | Пример |
|------------|----------|--------|
| `$S2T_FILE_PATH` | **Путь к файлу** | `/tmp/shell2telegram_files/1234_doc.txt` |
| `$S2T_LOGIN` | Telegram username | `john_doe` |
| `$S2T_USERID` | User ID | `123456789` |
| `$S2T_USERNAME` | Полное имя | `John Doe` |
| `$S2T_CHATID` | Chat ID | `123456789` |

---

## Обратная совместимость

Если НЕ определена команда `/:file`, то файлы будут обрабатываться через `/:plain_text` (старое поведение).

```bash
# Старый способ (все еще работает):
./shell2telegram_darwin_arm64 \
  /:plain_text 'bash -c "
if [ -n \"\$S2T_FILE_PATH\" ]; then
  echo \"Файл: \$S2T_FILE_PATH\"
else
  echo \"Текст: \$*\"
fi
"'

# Новый способ (рекомендуется):
./shell2telegram_darwin_arm64 \
  /:plain_text 'echo "Текст: $*"' \
  /:file 'echo "Файл: $(basename $S2T_FILE_PATH)"'
```

---

## FAQ

**Q: Что если определены И `/:plain_text` И `/:file`?**
A: Файлы → `/:file`, текст → `/:plain_text`

**Q: Что если определена только `/:file`?**
A: Файлы → `/:file`, текст → игнорируется

**Q: Как обработать файл повторно?**
A: Файлы сохраняются 1 час в `/tmp/shell2telegram_files/`, можешь создать команду для повторной обработки

**Q: Можно ли комбинировать с модификаторами?**
A: Да! `/:file:desc="My handler":md` работает

---

## Миграция со старого кода

Было:
```bash
/:plain_text 'bash -c "
if [ -n \"\$S2T_FILE_PATH\" ]; then
  # логика для файлов
else
  # логика для текста
fi
"'
```

Стало:
```bash
/:plain_text 'bash -c "# логика для текста"' \
/:file 'bash -c "# логика для файлов"'
```

---

**Теперь у тебя чистое разделение обработки файлов и текста! 🎉**
