@echo off
REM =============================================================================
REM start_bot.bat - Запуск Telegram бота shell2telegram с поддержкой прокси
REM =============================================================================
REM
REM Инструкция:
REM 1. Измените значения переменных ниже на свои
REM 2. Установите права доступа (запустите от Администратора):
REM    icacls start_bot.bat /inheritance:r /grant:r Administrators:F
REM 3. Запустите этот файл
REM
REM =============================================================================

REM ===== НАСТРОЙКА ПРОКСИ =====
REM Укажите адрес вашего прокси-сервера (например: proxy.company.com:8080)
set PROXY_SERVER=proxy.company.com:8080

REM Укажите имя пользователя для прокси
REM Для доменного пользователя: DOMAIN\username
set PROXY_USER=username

REM Укажите пароль для прокси
set PROXY_PASSWORD=your_password

REM ===== НАСТРОЙКА TELEGRAM БОТА =====
REM Получите токен от @BotFather в Telegram
set TB_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz_CHANGE_ME

REM ===== НАСТРОЙКА ROOT-ПОЛЬЗОВАТЕЛЕЙ =====
REM Укажите ваш Telegram username (без @)
set ROOT_USERS=admin

REM =============================================================================
REM ЗАПУСК БОТА
REM =============================================================================

echo ========================================
echo Запуск Telegram бота shell2telegram
echo ========================================
echo.
echo Прокси: %PROXY_SERVER%
echo Пользователь прокси: %PROXY_USER%
echo Root-пользователи: %ROOT_USERS%
echo.
echo Нажмите Ctrl+C для остановки бота
echo ========================================
echo.

REM Запуск бота с командами для Windows Server
REM ВАЖНО: Переменные окружения PROXY_SERVER, PROXY_USER, PROXY_PASSWORD и TB_TOKEN
REM        используются АВТОМАТИЧЕСКИ! Не нужно передавать их через параметры!
shell2telegram.exe ^
    -log-commands ^
    -log=bot.log ^
    -persistent-users ^
    -root-users=%ROOT_USERS% ^
    -add-exit ^
    -description="Бот для управления сервером" ^
    /help:desc="Помощь" "echo Используйте /help для списка команд" ^
    /status:desc="📊 Статус системы" "systeminfo | findstr /C:\"System Up Time\" /C:\"System Boot Time\"" ^
    /cpu:desc="💻 Загрузка CPU" "wmic cpu get loadpercentage /value" ^
    /memory:desc="🧠 Память" "systeminfo | findstr /C:\"Available Physical Memory\" /C:\"Total Physical Memory\"" ^
    /disk:desc="💾 Диски" "wmic logicaldisk get caption,freespace,size /format:table" ^
    /services:desc="⚙️ Службы" "net start" ^
    /processes:desc="📋 Процессы" "tasklist | sort /R /+65" ^
    /date:desc="📅 Дата" "date /t" ^
    /time:desc="🕐 Время" "time /t" ^
    /ping:vars=HOST:desc="🌐 Пинг" "ping -n 4 %HOST%" ^
    /ipconfig:desc="🔌 Сеть" "ipconfig /all" ^
    /uptime:desc="⏱️ Время работы" "systeminfo | findstr /C:\"System Boot Time\""

REM Если бот завершился, показываем сообщение
echo.
echo ========================================
echo Бот остановлен
echo ========================================
pause
