# Usar uma imagem base do Golang
FROM golang:1.22

# Definir o diretório de trabalho
WORKDIR /

# Copiar o go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código-fonte da aplicação
COPY . .

# Compilar a aplicação
RUN go build -o main .

# Definir o comando de entrada para o contêiner
CMD ["./main"]