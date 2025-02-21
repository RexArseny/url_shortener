URL shortener
---
Расчет общего тестового покрытия:
```
go test --coverprofile=coverage.out ./...
go tool cover --func=coverage.out
```
---
Профилирование потребление памяти:
```
curl http://localhost:8080/debug/pprof/heap -o profiles/base.pprof
go tool pprof -http=localhost:8080 profiles/base.pprof
```
---
Генерация документации:
```
godoc -http :8080
http://localhost:8080/pkg/?m=all
```