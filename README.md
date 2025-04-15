## Запуск сервера

В директории проекта:

> docker compose up pvz-service

Запуск тестов (77% coverage):

> docker compose up test

## Краткое описание
Для генерации type-safe запросов использовался sqlc, для генерации эндпоинтов по схеме - oapi-codegen

/api - сгенерированные эндпоинты

/internal/data - модель работы с БД и воркер функции

/internal/db - сгенерированное sqlc

/internal/handlers - хэндлеры и middleware

/internal/server - инициализация приложения

/interal/sql - схема БД и запросы для генерации c sqlc

