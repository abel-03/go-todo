# PlanPulse

Это полнофункциональное веб-приложение для управления списками дедлайнов или покупок.

## Функции   

Веб-приложение обеспечивает управление списком покупок до создания учетной записи с помощью веб-локального хранилища. При создании учетной записи он предлагает расширенные функции, такие как сохранение списков на разных устройствах и обмен списками покупок с другими пользователями.

## Используемые технологии

Бэкэнд написан на Go и использует библиотеку chi в качестве библиотеки маршрутизации. Он подключается к базе данных MongoDB для хранения пользователей и списков покупок.

Интерфейс построен с использованием TypeScript и React. Основные используемые библиотеки: Redux/rtk-query, Material UI и React Router. Статический пакет встроен в скомпилированный исполняемый файл Go и обслуживается из серверной части.

### Как запускать

Если вы хотите запустить этот проект локально, вы можете использовать следующие шаги:
1. Клонировать проект
2. Установите и запустите локальный экземпляр MongoDB.
3. Создайте файл .env со следующими переменными env.
```
MONGO_DB_URI="mongodb://localhost"
JWT_SIGN_KEY="your-secret-key-will-go-here"
PORT=8080
```
4. Установите godotenv ( https://github.com/joho/godotenv ) как команду bin. Он используется для предоставления переменных среды приложению. В качестве альтернативы вы можете реализовать другой способ предоставления этих переменных env.

5. В главном каталоге проекта введите:
```
make run
```
Это запустит скрипт в Makefile и запустит как бэкэнд, так и сервер разработки React. Сценарий сборки и запуска сначала скомпилирует проект, а затем запустит исполняемый файл.

MongoDB должен создать локальную базу данных при запуске приложения.