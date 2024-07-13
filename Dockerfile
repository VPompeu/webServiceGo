# Usar uma imagem base do Golang
FROM golang:1.22 as builder

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080

# Copiar o código-fonte da aplicação
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

FROM scratch
COPY --from=builder /app/server /server
ENTRYPOINT [ "/server"]