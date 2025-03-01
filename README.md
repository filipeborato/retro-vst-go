# RetroVST – README

## Descrição
O RetroVST é um projeto em Golang que fornece serviços de autenticação (incluindo login/password e Google OAuth), gerenciamento de usuários, pagamentos e transações, com um banco de dados SQLite. O projeto roda atrás do Nginx, e pode ser configurado para rodar via Docker ou localmente.

---
## Estrutura do Projeto

```bash
retro-vst-go/
├─ cmd/
│   ├─ main.go            # Ponto de entrada principal
│   └─ migrate_seed.go    # Script de migração e inserção de dados mock
├─ config/
│   └─ config.go          # Carrega variáveis de ambiente (JWT, etc.)
├─ db/
│   ├─ db.go              # Setup do GORM e SQLite
│   ├─ migration.go       # AutoMigrate + índices e triggers
│   └─ mock_data.go       # Inserção de dados de teste
├─ domain/
│   ├─ user.go            # Struct User
│   ├─ session.go         # Struct Session
│   ├─ payment.go         # Struct Payment
│   ├─ transaction.go     # Struct Transaction
│   └─ product.go         # Struct Product
├─ handlers/
│   ├─ auth_handler.go    # Lida com signup, login, Google OAuth
│   ├─ profile_handler.go # Rota de /profile
│   └─ ...
├─ nginx/
│   └─ default.conf       # Configuração do Nginx (proxy pass)
├─ docker-compose.yml     # Configuração Docker para rodar app + Nginx
├─ Dockerfile             # Build do app Go
├─ .env                   # Variáveis de ambiente (JWT_KEY, etc.)
├─ go.mod                 # Módulo Go
├─ go.sum                 # Checksums de dependências
└─ README.md              # Este arquivo
```

---
## Pré-Requisitos
- **Golang** v1.20 ou superior
- **SQLite** (bibliotecas necessárias se rodar local)
- **Nginx** (opcional, se rodar localmente sem Docker)
- **Docker & Docker Compose** (se for rodar via contêiner)
- **godotenv** (`github.com/joho/godotenv`) para carregar `.env`

---
## Configuração Local (Sem Docker)
1. Clone o repositório:
   ```bash
   git clone https://github.com/seu-usuario/retro-vst-go.git
   cd retro-vst-go
   ```

2. Crie um arquivo `.env` com as variáveis de ambiente:
   ```ini
   JWT_KEY=super_secret_jwt
   APP_PORT=8080
   SQLITE_DB_PATH=./db/retrovst.db
   APP_ENV=development
   ```

3. Instale dependências:
   ```bash
   go mod tidy
   ```

4. Rodar migração e dados de teste:
   ```bash
   go run cmd/migrate_seed.go
   ```
   - Ele vai perguntar se quer inserir mock data.

5. Executar o servidor:
   ```bash
   go run cmd/main.go
   ```
   - A aplicação estará escutando na porta `:8080` (ou `APP_PORT`).

6. Acesse as rotas de teste:
   - `GET /ping` => `{ "message": "pong" }`
   - `POST /signup` => Cria usuário
   - `POST /login` => Retorna JWT
   - `GET /api/profile` => Dados do usuário logado (JWT)

---
## Configuração via Docker

1. Crie (ou ajuste) seu `.env` com as variáveis:
   ```ini
   JWT_KEY=super_secret_jwt
   APP_PORT=8080
   SQLITE_DB_PATH=/app/db/retrovst.db
   APP_ENV=production
   ```

2. Build e subir contêineres:
   ```bash
   docker-compose up --build -d
   ```
   - Isso criará:
     - **retrovst** container: app Go + SQLite
     - **nginx-proxy** container: Nginx como proxy reverso

3. Verificar logs:
   ```bash
   docker-compose logs -f
   ```

4. Testar:
   - Acesse `http://localhost/` => Roteia para `retrovst` na porta 8080.
   - Se configurou `/retro-vst/`, acesse `http://localhost/retro-vst/`.

---
## Uso do Nginx
- O arquivo `nginx/default.conf` contém:
  ```nginx
  server {
      listen 80;
      server_name localhost;

      location / {
          proxy_pass http://retrovst:8080;
      }

      # HTTPS (exemplo)
      listen 443 ssl;
      ssl_certificate /etc/nginx/ssl/selfsigned.crt;
      ssl_certificate_key /etc/nginx/ssl/selfsigned.key;
  }
  ```
  Ajuste conforme seu domínio ou IP.

---
## Autenticação e JWT
- A aplicação usa JWT para rotas protegidas.
- Defina `JWT_KEY` no `.env` para ser a chave de assinatura.
- Chame `POST /login` para obter o token e inclua nos headers:
  ```http
  Authorization: Bearer <token>
  ```
- Rotas como `GET /api/profile` requerem esse token.

---
## Google OAuth
- Para habilitar Google OAuth, defina `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, etc. no `.env`.
- Rota de callback: `/auth/google/callback`.

---
## Referências
- [GORM](https://gorm.io)
- [Gin Gonic](https://github.com/gin-gonic/gin)
- [Docker Compose](https://docs.docker.com/compose/)
- [godotenv](https://github.com/joho/godotenv)

---
## Contribuição
- Faça um fork do repositório.
- Crie sua feature branch: `git checkout -b feature/minha-feature`
- Commit suas alterações: `git commit -m 'Minha nova feature'`
- Faça push da branch: `git push origin feature/minha-feature`
- Abra um Pull Request.

---
## Licença
Este projeto está sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para mais detalhes.

