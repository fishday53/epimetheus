# cmd/agent

В данной директории будет содержаться код Агента, который скомпилируется в бинарное приложение

Build example:
```bash
cd metrics-agent
go build -ldflags \
   "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse HEAD)'" \
   -o metrics-agent ./cmd/
```
