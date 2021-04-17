* NOTIFICATION SERVICE

Сервис уведомлений на golang с использованием gorilla/websocket и redis

* Требования:
+ Docker с docker-compose
+ Свободный порт 8080

* Инструкция по сборке и запуску:
1. Склонировать репозиторий: 

``` git clone https://github.com/neozumm/dpl_notifications ```

2. Убедиться, что docker запущен
3. Собрать через 
``` docker-compose up ```

Тестовый веб-интерфейс доступен на localhost:8080
Апи принимает POST запросы на подачу сообщений на localhost:8080/send в формате jsonrpc. 
Пример запроса: 

``` 
{"jsonrpc": "2.0", "method": "send", "params": {"msg": "test", "id": 111}, "id": 1}
```

На данный момент важный объект только params. Он принимает "msg" - текст сообщения и "id" - int ид клиента.
