# Usar uma imagem base do Golang
FROM golang:1.22 as builder

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código-fonte da aplicação
COPY . .

# Copiar o arquivo .env para dentro do contêiner
COPY .env .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Segunda etapa para construir a imagem final mínima
FROM scratch
COPY --from=builder /app/server /server
COPY --from=builder /app/.env /.env
ENTRYPOINT [ "/server" ]
