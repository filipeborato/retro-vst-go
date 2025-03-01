# RetroVST – README

## Description
RetroVST is a Golang project that provides authentication services (including login/password and Google OAuth), user management, payments and transactions, with a SQLite database. The project runs on top of Nginx, and can be configured to run via Docker or locally.

---
## Project Structure

```bash
retro-vst-go/
├─ cmd/
│ ├─ main.go # Main entry point
│ └─ migrate_seed.go # Migration script and mock data insertion
├─ config/
│ └─ config.go # Load environment variables (JWT, etc.)
├─ db/
│ ├─ db.go # GORM and SQLite setup
│ ├─ migration.go # AutoMigrate + indexes and triggers
│ └─ mock_data.go # Test data insertion
├─ domain/
│ ├─ user.go # Struct User
│ ├─ session.go # Struct Session
│ ├─ payment.go # Struct Payment
│ ├─ transaction.go # Struct Transaction
│ └─ product.go # Struct Product
├─ handlers/
│ ├─ auth_handler.go # Handles signup, login, Google OAuth
│ ├─ profile_handler.go # /profile route
│ └─ ...
├─ nginx/
│ └─ default.conf # Nginx configuration (proxy pass)
├─ docker-compose.yml # Docker configuration to run app + Nginx
├─ Dockerfile # Build of the Go app
├─ .env # Environment variables (JWT_KEY, etc.)
├─ go.mod # Go module
├─ go.sum # Dependency checksums
└─ README.md # This file
```

---
## Prerequisites
- **Golang** v1.20 or higher
- **SQLite** (required libraries if running locally)
- **Nginx** (optional, if running locally without Docker)
- **Docker & Docker Compose** (if running via container)
- **godotenv** (`github.com/joho/godotenv`) to load `.env`

---
## Local Configuration (Without Docker)
1. Clone the repository:
   
```bash
git clone https://github.com/your-username/retro-vst-go.git
cd retro-vst-go
```

2. Create a `.env` file with the environment variables:
 ```ini
JWT_KEY=super_secret_jwt
APP_PORT=8080
SQLITE_DB_PATH=./db/retrovst.db
APP_ENV=development
```

3. Install dependencies:

```bash
go mod tidy
```

4. Run migration and test data:
```bash
go run cmd/migrate_seed.go
```
- It will ask if you want to insert mock data.

5. Run the server:
```bash
go run cmd/main.go
```
- The application will be listening on port `:8080` (or `APP_PORT`).

6. Access the test routes:
- `GET /ping` => `{ "message": "pong" }`
- `POST /signup` => Create user
- `POST /login` => Returns JWT
- `GET /api/profile` => Logged in user data (JWT)

---
## Configuration via Docker

1. Create (or adjust) your `.env` with the variables:
```ini
JWT_KEY=super_secret_jwt
APP_PORT=8080
SQLITE_DB_PATH=/app/db/retrovst.db
APP_ENV=production
```

2. Build and start containers:
```bash
docker-compose up --build -d
```
- This will create:
- **retrovst** container: Go app + SQLite
- **nginx-proxy** container: Nginx as reverse proxy

3. Verify logs:
```bash
docker-compose logs -f
```

4. Test:
- Access `http://localhost/` => Route to `retrovst` on port 8080.
- If you configured `/retro-vst/`, access `http://localhost/retro-vst/`.

---
## Using Nginx
- The `nginx/default.conf` file contains:
```nginx
server {
listen 80;
server_name localhost;
location / {
proxy_pass http://retrovst:8080;
}

# HTTPS (example)
listen 443 ssl;
ssl_certificate /etc/nginx/ssl/selfsigned.crt; ssl_certificate_key /etc/nginx/ssl/selfsigned.key;
}
```
Adjust according to your domain or IP.

---
## Authentication and JWT
- The application uses JWT for protected routes.
- Set `JWT_KEY` in `.env` to be the signing key.
- Call `POST /login` to get the token and include in the headers:
```http
Authorization: Bearer <token>
```
- Routes like `GET /api/profile` require this token.

---
## Google OAuth
- To enable Google OAuth, set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, etc. in `.env`.
- Callback route: `/auth/google/callback`.

---
## References
- [GORM](https://gorm.io)
- [Gin Gonic](https://github.com/gin-gonic/gin)
- [Docker Compose](https://docs.docker.com/compose/)
- [godotenv](https://github.com/joho/godotenv)

---
## Contribute
- Fork the repository.
- Create your feature branch: `git checkout -b feature/my-feature`
- Commit your changes: `git commit -m 'My new feature'`
- Push the branch: `git push origin feature/my-feature`
- Open a Pull Request.

---
## License
This project is licensed under the MIT license - see the [LICENSE] file for more details.
