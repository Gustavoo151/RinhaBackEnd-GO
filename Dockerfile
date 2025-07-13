FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiando arquivos necessários
COPY go.mod go.sum ./
RUN go mod download

# Copiando o código fonte
COPY . .

# Compilando a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment-intermediator .

# Imagem final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiando o binário compilado
COPY --from=builder /app/payment-intermediator .

# Expondo a porta da API
EXPOSE 9999

# Executando a aplicação
CMD ["./payment-intermediator"]