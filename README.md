# gRPCServer

gRPC сервер, у которого есть метод `getReasonOfAbsence` для получения причины отсутствия человека, по его контактным данным. Сервер делает запрос к внешнему http-серверу и получает необходимые данные. 
В основе обработки запросов лежит пул обработчиков, который можно указать в конфигурации. Запросы к внешнему серверу кешируются. Есть логирование в разные файлы. Функционал приложения покрыт Unit-тестами с моками.


## RPC-метод сервера:

```protobuf
/**
 * Messages related to modification of the personal information.
 *
 */
syntax = "proto3";

package dataModification;

option go_package = "dataModification";

/**
 * Represents the personal information.
 */
message ContactDetails {
  string displayName = 1; /// full name of a person (first name, middle name, last name)
  string email = 2; /// email of person, which is an identifier
  string mobilePhone = 3; /// mobile phone number of person
  string workPhone = 4; /// work phone number of person
}

/**
 * Service for modifying the information of a specific person.
 */
service PersonalInfo {
  /// Used get the reason of absence of a specific person. Pass in a ContactDetails and modified ContactDetails will be returned.
  rpc getReasonOfAbsence(ContactDetails) returns (ContactDetails);
}

```
**Более визуальная документация: [Ссылка на документацию к proto файлу](./doc/personal_info_proto.html)**

## Конфигурация:
В основе лежит json структура, где все поля обязательные. Самое главное нужно не забыть указать два пути на внешний сервер.

```json
{
  "appServerInfo": {
    "serverIp": "127.0.0.1",
    "serverPort": "8080",
    "amountOfWorkers": 8,
    "queueSize": 10,
    "ttlOfItemsInCache": 20000,
    "logLevel": "DEBUG"
  },

  "externalServerInfo": {
    "employeeUrlPath": "https://127.0.0.1:8081/Portal/springApi/api/employees",
    "absenceUrlPath": "https://127.0.0.1:8081/Portal/springApi/api/absences",
    "requestTimeout": 2000,
    "login": "somelogin",
    "password": "password"
  }
}
```

**Более визуальная документация с ограничениями полей: [Ссылка на документацию к конфиг файлу](./doc/config_doc.html)**

Для запуска приложения:
```
make run
```

Для тестирования приложения: 
```
make test
```

