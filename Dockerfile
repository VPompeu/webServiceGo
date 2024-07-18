# Primeira etapa: compilar a aplicação
FROM golang:1.22 AS builder

WORKDIR /app

# Copiar o go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código-fonte da aplicação
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server .

# Segunda etapa: construir a imagem final mínima
FROM gcr.io/distroless/base-debian10

WORKDIR /app

# Copiar o executável compilado
COPY --from=builder /server .

# Copiar o arquivo .env para dentro do contêiner
COPY .env .

# Expor a porta utilizada pela aplicação
EXPOSE 8080

# Definir o comando de inicialização
ENTRYPOINT [ "./server" ]
