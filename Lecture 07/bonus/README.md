# Starter СР-2

Реализуйте TODO в пакетах `service` и `repository`. Публичные имена и сигнатуры
менять нельзя: их использует преподавательский набор тестов.

```bash
go test ./...
go test -cover ./...
gofmt -w .
```

Перед сдачей опишите в этом README направление зависимостей между пакетами и
приведите итоговый процент покрытия бизнес-логики.

## Описание решения

Сейчас проект имеет следующее направление зависимостей:
`repository -> service -> model`

- `service` использует `model`, чтобы работать с `model.Order`
- `repository` импортирует `service`, чтбы использовать `DeilveryPricer`, `ErrUnknownTariff` и конструкторы тарифов в `Default()`
- `model` не зависит от других пакетов

`Calculator` использует интерфейс `TariffRepository`, который объявлен внутри `service`, поэтому `service` не импортирует `repository`

Процент покрытия для бизнес логики (`service`) с помощью `go test -cover ./...` составляет 61.1%

### P.S.

Мне кажется странным, что бизнес сценарий (`Calculator`), тарифы и ошибки объединяются в `service`. Из-за этого `repository` зависит от `service`
Также странно, что `repository` имеет конструктор `Default()`, хотя по идее выбор конкретных компонентов это задача точки входа в программу

Чтобы распутать архитектуру я бы сделал следующее:

- Вынес `DeliveryPricer`, `Courier`, `Pickup`, `Express`, `BaseTariff` и ошибки `ErrUnknownTariff`, `ErrInvalidWeight`, `ErrInvalidDistance`, `ErrDistanceTooLarge` в отдельный пакет `tariff`
- Убрал бы `Default()` из `repository`
- Оставил `TariffRepository` в `service`, так как его использует именно `Calculator()`

Тогда у проекта будут понятные зависимости:

`service` -> `tariff`, `model`

`repositoty` -> `tariff`

`main` -> `service`, `repository`, `tariff`

Единственное, что мне остается не очень понятным: что делать с `DeliveryPricer`, потому что как я понимаю в Go интерфейсы располагают на стороне клиента, а не провайдера, но в моем варианте он в `tariff`

Может, я переношу свое видение с C++, где для меня это выглядит более явно

Интересно послушать, что Вы думаете по этому поводу
