# link-shortener

Link Shortener - сокращатель ссылок

## О проекте
Проект создан на Go и использует gRPC для авторизации пользователей  
В проекте реализованы следующие сервисы:
- URL Shortener - создание и редирект коротких ссылок
- SSO - управление пользователями и авторизация

## Требования
- Go 1.23.4
- Git

## Запуск проекта
1. Клонируйте репозиторий:  
```bash
git clone https://github.com/lostmyescape/link-shortener.git
```
2. Обновите зависимости:  
```bash
go mod tidy
```
3. Запуск проекта с Docker Compose:
