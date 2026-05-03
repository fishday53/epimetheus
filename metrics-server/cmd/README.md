# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение

```bash
curl -XPOST http://localhost:8080/update/gauge/a/1.53
curl http://localhost:8080/value/gauge/a
curl -XPOST http://localhost:8080/update/counter/b/-1
curl http://localhost:8080/value/counter/b
```

Build example:
```bash
cd metrics-server
go build -ldflags \
   "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse HEAD)'" \
   -o metrics-server ./cmd/
```

