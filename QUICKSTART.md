# 🚀 Быстрый старт: shell2telegram с автоматической поддержкой прокси

## ⚡ Как это работает

**shell2telegram АВТОМАТИЧЕСКИ читает переменные окружения!**

Вам НЕ нужно передавать `-proxy-server`, `-proxy-user`, `-proxy-password` в параметрах командной строки.

Просто установите переменные окружения, и программа их найдет сама!

## 📝 Простейший пример (Windows)

```cmd
REM 1. Установите переменные окружения
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=your_password
set TB_TOKEN=your_bot_token

REM 2. Запустите бота (без указания прокси в параметрах!)
shell2telegram.exe /date "date /t" /time "time /t"
```

**Всё!** Программа автоматически использует `PROXY_SERVER`, `PROXY_USER` и `PROXY_PASSWORD`.

## 🔍 Как программа находит настройки прокси

Программа проверяет переменные окружения в таком порядке:

### 1. Сначала проверяются параметры командной строки (высший приоритет)
```cmd
shell2telegram.exe -proxy-server proxy:8080 -proxy-user user -proxy-password pass /date "date /t"
```

### 2. Если параметры не указаны, проверяются кастомные переменные
```cmd
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=password
```

### 3. Если кастомные переменные не найдены, проверяются стандартные
```cmd
set HTTP_PROXY=http://username:password@proxy:8080
REM или
set HTTP_PROXY=http://proxy:8080
set PROXY_USER=username
set PROXY_PASSWORD=password
```

### 4. Поддерживаются также (в порядке приоритета)
- `PROXY_SERVER` (кастомная, для этого проекта)
- `HTTP_PROXY` (стандартная, uppercase)
- `http_proxy` (стандартная, lowercase - для Linux)
- `HTTPS_PROXY` (стандартная, uppercase)
- `https_proxy` (стандартная, lowercase - для Linux)

## 💡 Примеры использования

### Пример 1: Минимальная конфигурация

**config.bat:**
```cmd
@echo off
set PROXY_SERVER=proxy:8080
set PROXY_USER=user
set PROXY_PASSWORD=pass
set TB_TOKEN=your_token
```

**Запуск:**
```cmd
call config.bat
shell2telegram.exe /date "date /t"
```

### Пример 2: Использование HTTP_PROXY (совместимость)

```cmd
set HTTP_PROXY=http://user:pass@proxy:8080
set TB_TOKEN=your_token
shell2telegram.exe /uptime "systeminfo | findstr Boot"
```

### Пример 3: Переопределение через параметры

```cmd
REM Установлены переменные окружения
set PROXY_SERVER=old-proxy:8080
set TB_TOKEN=your_token

REM Но можно переопределить через параметр командной строки
shell2telegram.exe -proxy-server new-proxy:3128 /date "date /t"
```
В этом случае будет использован `new-proxy:3128`, а не `old-proxy:8080`.

### Пример 4: Linux/Mac

```bash
#!/bin/bash

# Устанавливаем переменные
export PROXY_SERVER=proxy.company.com:8080
export PROXY_USER=username
export PROXY_PASSWORD=password
export TB_TOKEN=your_token

# Запускаем (прокси используется автоматически!)
./shell2telegram /date 'date' /uptime 'uptime'
```

## 🔒 Безопасность

### ✅ ПРАВИЛЬНО (переменные окружения)
```cmd
set PROXY_PASSWORD=secret
shell2telegram.exe /date "date /t"
```
Пароль НЕ виден в списке процессов.

### ❌ НЕПРАВИЛЬНО (параметры командной строки)
```cmd
shell2telegram.exe -proxy-password secret /date "date /t"
```
Пароль ВИДЕН в списке процессов (`tasklist` / `ps aux`).

## 📊 Таблица приоритетов

| Способ настройки | Приоритет | Безопасность |
|------------------|-----------|--------------|
| `-proxy-server` (параметр) | 1 (высший) | ⚠️ Низкая (виден в процессах) |
| `PROXY_SERVER` (переменная) | 2 | ✅ Высокая |
| `HTTP_PROXY` (переменная) | 3 | ✅ Высокая |
| `http_proxy` (переменная) | 4 | ✅ Высокая |

**Рекомендация:** Используйте переменные окружения вместо параметров командной строки.

## 🎯 Готовые примеры

В репозитории есть готовые файлы:

- **start_bot.bat** - Полнофункциональный бот для Windows Server
- **example_simple.bat** - Минимальный пример для тестирования

Просто отредактируйте переменные в начале файла и запустите!

## ❓ FAQ

**Q: Нужно ли указывать `-proxy-server`, если я установил `PROXY_SERVER`?**
A: НЕТ! Программа автоматически найдет переменную окружения.

**Q: Что если у меня уже установлена `HTTP_PROXY` для других программ?**
A: Отлично! shell2telegram автоматически использует её. Можно также добавить `PROXY_USER` и `PROXY_PASSWORD` отдельно.

**Q: Как проверить, какие переменные установлены?**
A: Windows: `set | findstr PROXY` или `set | findstr TB_TOKEN`
   Linux/Mac: `env | grep -i proxy` или `env | grep TB_TOKEN`

**Q: Можно ли указать учетные данные прямо в HTTP_PROXY?**
A: Да! Формат: `http://username:password@proxy:8080`

**Q: Работает ли это на Linux/Mac?**
A: Да! Все переменные окружения работают одинаково на всех платформах.

## 📖 Дополнительная документация

- **PROXY_README.md** - Полная документация по работе с прокси
- **README_RU.md** - Полное руководство на русском языке с примерами
- **README.md** - Оригинальная документация (английский)

---

**💡 Совет:** Для постоянной установки переменных в Windows используйте:
```cmd
setx PROXY_SERVER "proxy:8080" /M
setx PROXY_USER "username" /M
setx PROXY_PASSWORD "password" /M
```
(Требуются права Администратора, перезапустите cmd после установки)
