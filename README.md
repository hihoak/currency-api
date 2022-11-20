# currency-api

## Основные механизмы

1) `API` - REST api представлен 4 сущностями:
   1) `registrator` - регистрирует и авторизует пользователей, аппрувит заявки на регистрацию
   2) `timeliner` - отдает исторические данные по курсам валют
   3) `users` - получение, перечисление, блокировка пользователей
   4) `walleter` - добавление, списание, обмен денег. Создание, перечисление счетов. Перечисление транзакций совершенных пользователем
2) `Exchanger` - внутренний обменник, собирает раз в 10 секунд информацию о валютном рынке, необходим для того чтобы не
   перегружать запросами внешний API + ускоряет работу приложения. Он предоставаляет всю информаци для API
   по текущему курсу и составляет исторические данные в БД
3) `Quoter` - котировщик, клиент внешнего сервиса, который предоставляет котировки из внешнего мира.
   В нашем случае есть пока что только мок. Обращения к нему от Exchanger ограничены по времени по причинам выше
4) `Storager` - интерфейс для базы данных, в нашем случае Postgres. Развернуто в docker с помощью
   docker compose

## Конфигурация
Для конфигурации можно использовать переменные среды или конфиг файл переданный через флаг `--config`:
```yaml
   # service
   CURRENCY_API_SERVER_ADDRESS: "0.0.0.0:8000" # адрес по которому будет слушать сервис
   # database
   CURRENCY_API_DATABASE_HOST: "127.0.0.1" # хост от ДБ
   CURRENCY_API_DATABASE_PORT: "5432" # порт от ДБ
   CURRENCY_API_DATABASE_USER: "postgres"
   CURRENCY_API_DATABASE_PASSWORD: "password" # пароль от ДБ
   CURRENCY_API_DATABASE_CONNECTION_TIMEOUT: 2s # таймаут на коннекты к базе

   # logger
   CURRENCY_API_LOGGER_LOG_LEVEL: "debug"
```

```yaml
   server:
      address: "0.0.0.0:8000"
   database:
      host: "127.0.0.1"
      port: "5432"
      user: "postgres"
      password: "pass"
      connection_timeout: 2s
   logger:
      log_level: "debug"
```

## Архитектура

Архитектура состоит из 2 компонентов - БД Postgresql и Сервис на Golang. 
Весь контур поднимается командой `make up` с помощью docker compose

## Схема БД

Миграции находятся в папке `migrations` и накатываются с помощью команды `make migrate-up` и утилиты `goose`

<img width="651" alt="Screenshot 2022-11-20 at 11 53 53" src="https://user-images.githubusercontent.com/92049351/202895165-a8641776-3199-43b9-b361-0cbb20d4105a.png">

## API

### POST /register
```
POST /register - регистрирует нового пользователя

{
    "user": {
        "name" string,
        "middle_name": string,
        "mail": string,
        "phone_number": string,
        "password": string
    }
}
```

### /register/approve
```
POST /register/approve - подтверждает регистрацию пользователя

{
    "user": {
        "id": int64
    }
}
```

### /login
```
/login - выполняет авторизацию, необходимо хотя быодно из полей  phone_number или email

{
    "phone_number": string,
    "email": string,
    "password": string
}
```

### /user/block
```
POST /user/block - блокирует или разблокирует пользователя в зависимости от поля "block"

{
    "user": {
        "id": string,
    },
    "block": boolean
}
```

### /user/list
```
POST /user/list - перечисляет всех пользователей

offset, count - параметры пагинации

{
    "offset": int64,
    "count": int64
}
```

### /user/info
```
/user/info - выводит полную информацию о пользователе

request:
{
    "id": 1
}

response:
{
    "user" {
        "id": int64,
        "name": string
        "middle_name": string
        "surname": string
        "mail": string
        "phone_number": string
        "blocked" bool
        "registered" bool
        "admin" bool
        "password" string
    },
    
    "wallets": [
        {
            "id" int64
            "user_id" int64
            "currency" Currency
            "value" int64
        }
    ]
}
```

### /wallet/get
```
POST /wallet/get - достает конкретный счет по ID

{
    "id": 1
}
```

### /wallet/create
```
POST /wallet/create - создает счет

{
    "wallet": {
        "user_id": int64,
        "currency": Currency,
        "value": int64
    }
}
```

### /wallet/money/add
```
POST /wallet/money/add - добавляет указанное кол-во денег на счет

{
    "id": int64,
    "value": int64
}
```

### "/wallet/money/pull"
```
POST /wallet/money/pull - списывает указанное кол-во денег со счета

{
    "id": int64,
    "amount": int64
}
```

### /wallet/list
```
POST /wallet/list - перечисляет все счета для кокретного пользователя

{
    "user_id": int64
}
```

### /wallet/exchange
```
POST /wallet/exchange - основной метод обмена валют, меняет валюту для user_id
с кошелька from_wallet_id на кошелек to_wallet_id с типами валют соответственно
from_currency и to_currency на сумму amount

{
    "user_id": int64,
    "from_wallet_id": int64,
    "to_wallet_id": int64,
    "from_currency": Currency,
    "to_currency": Currency,
    "amount": int64
}
```

### /wallet/courses
```
POST /wallet/course - отдает текущий курс для USD и EUR

{
    "from": Currency,
    "to": Currency
}
```

### /transaction/list
```
POST /transaction/list - перечисляет все операции сделанные пользователем,
в них входят 
"ADD MONEY" - попоплнение баланса (ручка /wallet/money/add)
"PULL MONEY" - вывод с баланса  (ручка /wallet/money/pull)
"EXCHANGE MONEY" - обмен валют (ручка /wallet/exchange)

{
    "user_id": int64
}
```

### /currency/list
```
POST /currency/list - отдает список всех валют доступных системой
текущий список RUB, EUR, USD, GBP, JPY, CHF, CNY

{}
```


### /course/list
```
POST /course/list - позволяет достать исторические данные из локальной базы данных
используется для отрисовки графиков 

{
    "from": Currency,
    "to": Currency,
    "from_time": int64,
    "to_time": int64
}
```

