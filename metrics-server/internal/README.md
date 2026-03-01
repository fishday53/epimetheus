# internal

В данной директории и её поддиректориях будет содержаться имплементация вашего сервиса

```bash
export DATABASE_DSN="user=myuser password=******** host=localhost port=5432 dbname=mydb sslmode=disable"
```
```bash
curl -X POST -H "Content-Type: application/json" -d '{"id":"x","type":"counter","delta":1}' http://localhost:8080/update/
```
```bash
curl -X POST -H "Content-Type: application/json" -d '
[
  {
    "id": "a",
    "type": "gauge",
    "value": 154672
  },
  {
    "id": "aa",
    "type": "counter",
    "delta": 1
  },
  {
    "id": "a",
    "type": "gauge",
    "value": 0.5737458002440649
  },
  {
    "id": "aa",
    "type": "counter",
    "delta": 2
  }
]' http://localhost:8080/updates/
```
