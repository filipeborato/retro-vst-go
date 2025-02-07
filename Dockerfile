# 1. Usa a imagem oficial do Golang como base
FROM golang:1.20-alpine

# 2. Define o diretório de trabalho no container
WORKDIR /app

# 3. Instala as dependências do SQLite (importante para CGO)
RUN apk add --no-cache gcc musl-dev

# 4. Copia os arquivos do projeto para dentro do container
COPY . .

# 5. Habilita CGO para funcionar com SQLite
ENV CGO_ENABLED=1

# 6. Instala as dependências
RUN go mod tidy

# 7. Compila a aplicação
RUN go build -o main ./cmd/main.go

# 8. Executa a migração do banco antes de rodar a aplicação
RUN go run cmd/migrate_seed.go

# 9. Expõe a porta definida nas variáveis de ambiente
EXPOSE ${APP_PORT}

# 10. Define o comando para rodar o app
CMD ["./main"]
