# Facilitador de Doações

<img width="1580" height="861" alt="Facilitador de Doações - Arquitetura (2)" src="https://github.com/user-attachments/assets/973f67cc-70c2-473e-83be-6b12cafdf5cf" />

Plataforma para conectar doadores a instituições e campanhas sociais. Permite realizar doações via PIX ou cartão de crédito, acompanhar campanhas em andamento, e gerenciar as necessidades de cada instituição.

> Este repositório contém apenas o backend (API). O frontend está disponível em: [link do repositório do frontend](https://github.com/jufazenda/facilitador-doacoes-front)

---

## O que esse projeto faz

- Usuários podem se cadastrar, buscar instituições e campanhas, e realizar doações
- Instituições se cadastram, criam campanhas e publicam suas necessidades
- Pagamentos são processados via ASAAS (PIX e cartão de crédito)
- Imagens de perfil e de instituições são armazenadas no Supabase
- A autenticação é feita via Auth0 com tokens JWT
- Há um sistema de ranking dos doadores mais ativos

---

## Tecnologias

- **Go** com Gin (web framework)
- **PostgreSQL** (banco de dados)
- **GORM** (ORM)
- **Auth0** (autenticação e controle de acesso)
- **ASAAS** (gateway de pagamento)
- **Supabase** (armazenamento de arquivos)
- **Redis + Asynq** (fila de tarefas assíncronas para notificações)
- **Docker** (banco de dados em desenvolvimento)

---

## Pré-requisitos

Para rodar o projeto localmente você vai precisar de:

- [Go 1.22+](https://golang.org/dl/)
- [Docker](https://www.docker.com/) — para subir o banco de dados localmente
- [Make](https://www.gnu.org/software/make/)
- Uma conta no [Auth0](https://auth0.com/) com uma aplicação e a Management API configuradas
- Uma conta no [Supabase](https://supabase.com/) com um bucket criado
- Uma conta no [ASAAS](https://www.asaas.com/) (pode usar o sandbox para testes)
- Redis (opcional — só necessário para o envio de notificações assíncronas)

---

## Configuração

Copie o arquivo de exemplo e preencha com suas credenciais:

```bash
cp .env.example .env
```

### Variáveis de ambiente

**Banco de dados**

| Variável      | Descrição     | Exemplo     |
| ------------- | ------------- | ----------- |
| `DB_HOST`     | Host do banco | `localhost` |
| `DB_USER`     | Usuário       | `admin`     |
| `DB_PASSWORD` | Senha         | `secret`    |
| `DB_NAME`     | Nome do banco | `app`       |
| `DB_PORT`     | Porta         | `5432`      |
| `DB_SSLMODE`  | Modo SSL      | `disable`   |

**Servidor**

| Variável       | Descrição                       | Exemplo                 |
| -------------- | ------------------------------- | ----------------------- |
| `PORT`         | Porta da API                    | `8080`                  |
| `FRONTEND_URL` | URL do frontend (usada no CORS) | `http://localhost:5174` |

**Auth0**

| Variável                   | Descrição                                          |
| -------------------------- | -------------------------------------------------- |
| `AUTH0_DOMAIN`             | Domínio do seu tenant (ex: `seu-tenant.auth0.com`) |
| `AUTH0_AUDIENCE`           | Audience configurada na API do Auth0               |
| `AUTH0_MGMT_CLIENT_ID`     | Client ID da Management API                        |
| `AUTH0_MGMT_CLIENT_SECRET` | Client Secret da Management API                    |

**ASAAS (pagamentos)**

| Variável              | Descrição                                 |
| --------------------- | ----------------------------------------- |
| `ASAAS_API_KEY`       | Chave de API                              |
| `ASAAS_SANDBOX`       | `true` para testes, `false` para produção |
| `ASAAS_WEBHOOK_TOKEN` | Token de validação dos webhooks           |

**Supabase (armazenamento)**

| Variável               | Descrição                                         |
| ---------------------- | ------------------------------------------------- |
| `SUPABASE_URL`         | URL do projeto (ex: `https://xyzxyz.supabase.co`) |
| `SUPABASE_KEY`         | Service role key                                  |
| `SUPABASE_BUCKET_NAME` | Nome do bucket                                    |

**Redis (opcional)**

| Variável             | Descrição                      | Exemplo                        |
| -------------------- | ------------------------------ | ------------------------------ |
| `REDIS_URL`          | URL de conexão                 | `redis://localhost:6379`       |
| `CHATBOT_NOTIFY_URL` | URL para envio de notificações | `http://localhost:3000/notify` |

---

## Como rodar

**1. Suba o banco de dados:**

```bash
make db/up
# ou diretamente:
docker compose up -d
```

Isso inicia um container PostgreSQL via Docker. As tabelas são criadas automaticamente quando a aplicação sobe (auto-migrate).

**2. Rode a aplicação:**

```bash
make run
# ou diretamente:
go run ./cmd/api
```

A API ficará disponível em `http://localhost:8080`.

**Outros comandos:**

| Ação                   | Make           | Docker / Go                     |
| ---------------------- | -------------- | ------------------------------- |
| Subir o banco          | `make db/up`   | `docker compose up -d`          |
| Parar o banco          | `make db/down` | `docker compose down`           |
| Rodar a aplicação      | `make run`     | `go run ./cmd/api`              |
| Gerar binário          | `make build`   | `go build -o bin/api ./cmd/api` |
| Atualizar dependências | `make tidy`    | `go mod tidy`                   |

**Rodar os testes:**

```bash
go test ./...
```

---

## Endpoints

Base path: `/api/v1`

As rotas marcadas com **auth** exigem um token JWT válido no header:

```
Authorization: Bearer <token>
```

As marcadas com **instituição** exigem que o token pertença a uma instituição cadastrada. As marcadas com **usuário** exigem que o token pertença a um usuário cadastrado.

### Usuários

| Método   | Rota                | Acesso         | Descrição                           |
| -------- | ------------------- | -------------- | ----------------------------------- |
| `GET`    | `/users`            | público        | Lista todos os usuários             |
| `GET`    | `/users/:id`        | público        | Busca usuário por ID                |
| `GET`    | `/users/me`         | auth (usuário) | Retorna o usuário autenticado       |
| `POST`   | `/users`            | auth           | Cria um usuário                     |
| `PUT`    | `/users/:id`        | auth (usuário) | Atualiza dados do usuário           |
| `DELETE` | `/users/:id`        | auth (usuário) | Remove o usuário                    |
| `POST`   | `/users/:id/avatar` | auth (usuário) | Upload de foto de perfil (máx 20MB) |

### Instituições

| Método   | Rota                       | Acesso             | Descrição                                                               |
| -------- | -------------------------- | ------------------ | ----------------------------------------------------------------------- |
| `GET`    | `/institutions`            | público            | Lista todas as instituições                                             |
| `GET`    | `/institutions/:id`        | público            | Busca instituição por ID                                                |
| `GET`    | `/institutions/me`         | auth (instituição) | Retorna a instituição autenticada                                       |
| `POST`   | `/institutions`            | auth               | Cria uma instituição                                                    |
| `PUT`    | `/institutions/:id`        | auth (instituição) | Atualiza dados da instituição                                           |
| `DELETE` | `/institutions/:id`        | auth (instituição) | Remove a instituição                                                    |
| `PATCH`  | `/institutions/:id/status` | auth               | Aprova ou rejeita uma instituição (`pending` / `approved` / `rejected`) |
| `PATCH`  | `/institutions/:id/images` | auth (instituição) | Upload de imagem (`?type=logo` ou `?type=cover`)                        |

### Campanhas

| Método   | Rota                          | Acesso             | Descrição                                                         |
| -------- | ----------------------------- | ------------------ | ----------------------------------------------------------------- |
| `GET`    | `/campaigns`                  | público            | Lista campanhas (filtros: `?keyword=` e `?is_urgent=true`)        |
| `GET`    | `/campaigns/:id`              | público            | Busca campanha por ID                                             |
| `GET`    | `/institutions/:id/campaigns` | público            | Lista campanhas de uma instituição                                |
| `POST`   | `/institutions/:id/campaigns` | auth (instituição) | Cria uma campanha                                                 |
| `PUT`    | `/campaigns/:id`              | auth (instituição) | Atualiza uma campanha                                             |
| `DELETE` | `/campaigns/:id`              | auth (instituição) | Remove uma campanha                                               |
| `PATCH`  | `/campaigns/:id/status`       | auth (instituição) | Atualiza status (`active` / `paused` / `completed` / `cancelled`) |

### Necessidades

| Método   | Rota                            | Acesso             | Descrição                             |
| -------- | ------------------------------- | ------------------ | ------------------------------------- |
| `GET`    | `/institutions/:id/necessities` | público            | Lista necessidades de uma instituição |
| `GET`    | `/necessities/:id`              | público            | Busca necessidade por ID              |
| `POST`   | `/institutions/:id/necessities` | auth (instituição) | Cria uma necessidade                  |
| `PUT`    | `/necessities/:id`              | auth (instituição) | Atualiza uma necessidade              |
| `DELETE` | `/necessities/:id`              | auth (instituição) | Remove uma necessidade                |
| `PATCH`  | `/necessities/:id/status`       | auth (instituição) | Marca como atendida                   |

### Doações

| Método | Rota                 | Acesso  | Descrição                                                |
| ------ | -------------------- | ------- | -------------------------------------------------------- |
| `GET`  | `/donations`         | público | Lista todas as doações                                   |
| `GET`  | `/donations/:id`     | público | Busca doação por ID                                      |
| `POST` | `/donations`         | público | Cria uma doação (PIX ou cartão de crédito)               |
| `POST` | `/donations/webhook` | público | Webhook do ASAAS para atualização de status de pagamento |

### Ranking e Transferências

| Método | Rota         | Acesso             | Descrição                                  |
| ------ | ------------ | ------------------ | ------------------------------------------ |
| `GET`  | `/ranking`   | público            | Ranking de doadores (`?limit=10`)          |
| `POST` | `/transfers` | auth (instituição) | Cria uma transferência bancária ou via PIX |

---

## Estrutura do projeto

```
.
├── cmd/api/          # entrypoint da aplicação
├── internal/
│   ├── handler/      # recebe as requisições HTTP e responde ao cliente
│   ├── usecase/      # regras de negócio
│   ├── repository/   # acesso ao banco de dados
│   ├── model/        # entidades do domínio
│   └── middleware/   # autenticação e controle de acesso
└── pkg/
    ├── asaas/        # integração com gateway de pagamento
    ├── auth0mgmt/    # integração com Auth0 Management API
    ├── supabase/     # integração com armazenamento de arquivos
    ├── database/     # conexão com PostgreSQL
    └── queue/        # fila assíncrona com Redis
```

O projeto segue o fluxo: `Handler → UseCase → Repository`. Cada camada só conhece a interface da camada abaixo dela, o que facilita os testes e a troca de implementações.
