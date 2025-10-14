@echo off
REM =============================================================================
REM Простой пример запуска бота с автоматическим использованием переменных окружения
REM =============================================================================

echo.
echo ===== Установка переменных окружения =====
echo.

REM Устанавливаем переменные окружения
set PROXY_SERVER=proxy.company.com:8080
set PROXY_USER=username
set PROXY_PASSWORD=your_password
set TB_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

echo ✓ PROXY_SERVER=%PROXY_SERVER%
echo ✓ PROXY_USER=%PROXY_USER%
echo ✓ PROXY_PASSWORD=****** (скрыт)
echo ✓ TB_TOKEN=****** (скрыт)
echo.

echo ===== Запуск бота =====
echo.
echo ВАЖНО: Программа АВТОМАТИЧЕСКИ использует переменные окружения!
echo        Не нужно передавать -proxy-server, -proxy-user, -proxy-password!
echo.
echo Программа проверит переменные в таком порядке:
echo   1. PROXY_SERVER (если не указан -proxy-server)
echo   2. HTTP_PROXY (если PROXY_SERVER не найден)
echo   3. http_proxy (если HTTP_PROXY не найден)
echo.

REM Запуск бота - переменные используются автоматически!
shell2telegram.exe ^
    -log-commands ^
    /date:desc="Текущая дата" "date /t" ^
    /time:desc="Текущее время" "time /t" ^
    /ping:vars=HOST:desc="Пинг хоста" "ping -n 4 %HOST%"

echo.
echo Бот остановлен
pause
