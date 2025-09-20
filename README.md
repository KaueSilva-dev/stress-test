# Stress Test

[![Go Report Card](https://goreportcard.com/badge/github.com/kauesilva/stress-test)](https://goreportcard.com/report/github.com/kauesilva/stress-test)
[![Build Status](https://github.com/kauesilva/stress-test/actions/workflows/go.yml/badge.svg)](https://github.com/kauesilva/stress-test/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/kauesilva/stress-test/graph/badge.svg)](https://codecov.io/gh/kauesilva/stress-test)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Uma ferramenta de linha de comando (CLI) para realizar testes de carga em serviços web, com um modo de Interface de Usuário de Terminal (TUI) para visualização em tempo real.

Este projeto foi desenvolvido como parte do desafio da Pós-Graduação Go Expert.

## ✨ Funcionalidades

- **Teste de Carga HTTP**: Execute um número configurável de requisições HTTP para um endpoint.
- **Concorrência Ajustável**: Defina quantos workers simultâneos farão as requisições.
- **Modo TUI**: Acompanhe o progresso do teste em tempo real com uma interface de terminal interativa.
- **Relatório Detalhado**: Ao final do teste, receba um sumário com:
  - Total de requisições feitas com sucesso.
  - Requisições com status `200 OK`.
  - Total de erros.
  - Distribuição de todos os códigos de status HTTP.
  - Tempo total de execução e Requisições Por Segundo (RPS).

## 🚀 Instalação

Certifique-se de ter o Go (versão 1.21 ou superior) instalado.

Você pode instalar a ferramenta diretamente com `go install`:

```bash
go install github.com/kauesilva/stress-test@latest
```

## Usage

A ferramenta pode ser executada em dois modos: relatório simples (padrão) ou TUI.

### Argumentos

*   `--url`: (Obrigatório) A URL do serviço a ser testado.
*   `--total`: O número total de requisições a serem feitas (padrão: 1).
*   `--concurrency`: O número de workers simultâneos (padrão: 1).
*   `--tui`: Ativa o modo de Interface de Usuário de Terminal (TUI).

### Exemplos

#### Execução Padrão

Execute 1000 requisições com 10 workers simultâneos e veja o relatório no final.

```bash
docker run --rm -it kauesilva/rate-limiter:latest run --concurrency 10 --requests 1000 --timeout 30s --url https://www.google.com.br
```

#### Modo TUI

Para uma visualização em tempo real do progresso, use a flag `--tui`.

```bash
docker run --rm -it kauesilva/rate-limiter:latest run --concurrency 10 --requests 1000 --timeout 30s --tui --url https://www.google.com.br
```

A saída será parecida com esta:

```
Loadtester (TUI)
URL: http://localhost:8080 | Total: 1000 | Concorrência: 10

█████████████████████████████▋          55%

Progresso: 550/1000
HTTP 200: 545
Tempo: 5.8s | RPS: 94.83

Distribuição de status:
  200: 545
  429: 5

Pressione q para sair.
```

## 🛠️ Desenvolvimento

### Pré-requisitos

- Go (versão 1.21+)

### Compilando a partir do código-fonte

1.  Clone o repositório:
    ```bash
    git clone https://github.com/kauesilva/stress-test.git
    cd stress-test
    ```

2.  Compile o projeto:
    ```bash
    go build -o stress-test ./cmd/stress-test
    ```

### Rodando os testes

```bash
go test ./...
```

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.
