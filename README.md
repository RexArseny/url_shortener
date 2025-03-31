URL shortener
---
Расчет общего тестового покрытия:
```
go test --coverprofile=coverage.out ./cmd/shortener ./internal/app ./internal/app/config ./internal/app/controllers ./internal/app/logger ./internal/app/middlewares ./internal/app/models ./internal/app/repository ./internal/app/routers ./internal/app/usecases
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
---
Генерация сертификата и закрытого ключа:
```
go run ./cmd/certificate_generator/certificate_generator.go
```
---
Генерация go кода на основе proto файла:
```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative .\internal\app\models\proto\dto.proto
```