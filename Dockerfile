# Usar uma imagem base do Golang
FROM golang:1.22

# Adiciona o Cloud SQL Proxy
ADD https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 /cloud_sql_proxy
RUN chmod +x /cloud_sql_proxy

# Definir o diretório de trabalho
WORKDIR /

# Copiar o go.mod e go.sum e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080

# Copiar o código-fonte da aplicação
COPY . .

# Compilar a aplicação
RUN go build -o main .

# Definir o comando de entrada para o contêiner
CMD ["/cloud_sql_proxy", "-dir=/cloudsql", "&", "./main"]
