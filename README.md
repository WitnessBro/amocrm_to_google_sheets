# amoCRM -> Google Sheets

Сервис отслеживает события в amoCRM и при срабатывании триггерного поля записывает данные сделки и контактов в Google Sheets.

## Что делает

- проверяет события в amoCRM по таймеру;
- отбирает события изменения поля `custom_field_642629_value_changed` со значением `1`;
- получает карточку сделки, контактов и компании;
- добавляет отдельную строку в Google Sheets для каждого контакта-слушателя.

Telegram-логика в проект не входит.

## Настройка

1. Скопируйте пример конфигурации:

   `cp configs/config.yaml.example configs/config.yaml`

2. Заполните `configs/config.yaml`:
   - `amo_subdomain`
   - `amo_api_key`
   - `google_table_id`
   - `google_credentials_path`

3. Положите JSON-ключ service account в путь из `google_credentials_path`.
4. Выдайте service account права редактора на Google-таблицу.

## Запуск

```bash
go mod tidy
go run ./cmd/main.go
```
