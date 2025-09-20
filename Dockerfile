# Etapa de build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Binário estático para imagem mínima (desative CGO)
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /app/loadtester .

# Etapa final (distroless ou scratch). Usaremos scratch para mínima.
FROM scratch

WORKDIR /
COPY --from=builder /app/loadtester /loadtester

# Usuário não-root (opcional): scratch não tem /etc/passwd. Se precisar, use alpine.
# ENTRYPOINT simples:
ENTRYPOINT ["/loadtester"]
CMD ["run", "--help"]