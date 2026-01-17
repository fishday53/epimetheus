# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение

```bash
curl -XPOST http://localhost:8080/update/gauge/a/1.53
curl http://localhost:8080/value/gauge/a
curl -XPOST http://localhost:8080/update/counter/b/-1
curl http://localhost:8080/value/counter/b
```