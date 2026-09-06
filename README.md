# Logger

## Установка
`go get gitlab.satel.org/rtuc-forks/aqa/logger.git/pkg/logger`

## Конфигурация

| Переменная | Описание | Значение по умолчанию |
|------------|----------|---------------------|
| `LOGGER_DST` | Назначение логов: `local` - только slog | `""` (OpenTelemetry) |
| `LOGGER_MIN_LEVEL` | Минимальный уровень логирования: DEBUG, INFO, WARN, ERROR | `INFO` |
| `LOG_FILE` | Путь к файлу логов | `""` |
| `STAND_DOMAIN` | Имя хоста/домена для идентификации сервиса | `-` |
