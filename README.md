# Darts Training App - Backend

Dies ist das Backend für die Darts Training App, entwickelt mit Go 1.21+ und dem Gin Framework.

## Technologien

- **Go 1.21+** - Programmiersprache
- **Gin** - HTTP Web Framework
- **PostgreSQL** - Datenbank
- **GORM** - ORM (Object-Relational Mapping)
- **Auth0** - Authentifizierung
- **JWT** - JSON Web Tokens
- **Docker** - Containerisierung

## API Endpoints

### Health Check
- `GET /health` - Service Status

### Authentifizierung
- `POST /api/auth/login` - Login mit Auth0
- `GET /api/auth/me` - Aktueller Benutzer (geschützt)

### Teams (CRUD)
- `GET /api/teams` - Alle Teams
- `POST /api/teams` - Team erstellen
- `GET /api/teams/:id` - Team Details
- `PUT /api/teams/:id` - Team aktualisieren
- `DELETE /api/teams/:id` - Team löschen
- `GET /api/teams/:id/players` - Team Spieler

### Spieler (CRUD)
- `GET /api/players` - Alle Spieler
- `POST /api/players` - Spieler erstellen
- `GET /api/players/:id` - Spieler Details
- `PUT /api/players/:id` - Spieler aktualisieren
- `DELETE /api/players/:id` - Spieler löschen
- `GET /api/players/team/:teamId` - Spieler pro Team
- `GET /api/players/me` - Aktueller Benutzer
- `POST /api/players/me` - Aktuellen Benutzer erstellen

### Training Sessions (CRUD)
- `GET /api/training-sessions` - Alle Training Sessions
- `POST /api/training-sessions` - Training erstellen
- `GET /api/training-sessions/:id` - Training Details
- `PUT /api/training-sessions/:id` - Training aktualisieren
- `DELETE /api/training-sessions/:id` - Training löschen
- `POST /api/training-sessions/:id/start` - Training starten
- `POST /api/training-sessions/:id/finish` - Training beenden
- `GET /api/training-sessions/:id/costs` - Kostenberechnung
- `POST /api/training-sessions/:id/players` - Spieler hinzufügen
- `DELETE /api/training-sessions/players/:playerId` - Spieler entfernen

### Spiele
- `GET /api/games/modes` - Spielmodi
- `GET /api/games/training/:sessionId` - Spiele pro Training
- `POST /api/games/training/:sessionId` - Spiel erstellen
- `POST /api/games/training/:sessionId/generate` - Spiele generieren
- `PUT /api/games/:id` - Spiel aktualisieren
- `DELETE /api/games/:id` - Spiel löschen

## Environment Variablen

Kopiere `.env.example` nach `.env` und passe die Werte an:

```bash
cp .env.example .env
```

Benötigte Variablen:
- `PORT` - Server Port (default: 8080)
- `DATABASE_URL` - PostgreSQL Verbindung
- `AUTH0_DOMAIN` - Auth0 Domain
- `AUTH0_CLIENT_ID` - Auth0 Client ID
- `JWT_SECRET` - JWT Secret
- `FRONTEND_URL` - Frontend URL für CORS

## Lokale Entwicklung

### 1. Go installieren
```bash
# Go 1.21+ erforderlich
go version
```

### 2. Dependencies installieren
```bash
go mod download
```

### 3. Datenbank starten
```bash
# PostgreSQL mit Docker
docker run --name darts-postgres -e POSTGRES_DB=darts_training -e POSTGRES_USER=darts_user -e POSTGRES_PASSWORD=your_password -p 5432:5432 -d postgres:13
```

### 4. Environment konfigurieren
```bash
# .env Datei anpassen
DATABASE_URL=postgres://darts_user:your_password@localhost:5432/darts_training?sslmode=disable
```

### 5. Server starten
```bash
go run cmd/server/main.go
```

## Docker Build

```bash
# Build Docker image
docker build -t darts-training-api .

# Run container
docker run -p 8080:8080 --env-file .env darts-training-api
```

## Deployment auf Render.com

### 1. Repository vorbereiten
- Backend Code auf GitHub pushen
- `render.yaml` Konfiguration vorhanden
- `Dockerfile` optimiert für Production

### 2. PostgreSQL Datenbank erstellen
1. Render Dashboard → New → PostgreSQL
2. Database Name: `darts-training-db`
3. Plan: Free
4. Region: Wählen
5. Create Database

### 3. Backend Web Service erstellen
1. Render Dashboard → New → Web Service
2. GitHub Repository verbinden
3. Runtime: Docker
4. Root Directory: `backend`
5. Plan: Free
6. Environment Variablen setzen:
   - `DATABASE_URL` (von PostgreSQL)
   - `JWT_SECRET` (auto-generate)
   - `AUTH0_DOMAIN`
   - `AUTH0_CLIENT_ID`
   - `FRONTEND_URL`

### 4. Deployment überprüfen
- Health Check: `https://your-api.onrender.com/health`
- Logs überprüfen
- Database Connection testen

## Datenbank Schema

### Tables
- `teams` - Mannschaften
- `players` - Spieler
- `game_modes` - Spielmodi
- `training_sessions` - Training Sessions
- `training_players` - Spieler pro Training
- `training_games` - Spiele pro Training

### Auto-Migration
Die Anwendung führt automatisch Datenbank-Migrationen durch und erstellt Default-Daten (Spielmodi).

## Fehlerbehebung

### Build Fehler
```bash
# Go Module Probleme
go mod tidy

# Dependencies aktualisieren
go get -u ./...
```

### Database Connection
```bash
# PostgreSQL Status prüfen
docker logs darts-postgres

# Connection testen
psql $DATABASE_URL
```

### CORS Probleme
Stelle sicher, dass `FRONTEND_URL` korrekt gesetzt ist:
```
FRONTEND_URL=https://your-frontend.onrender.com
```

## Production Tipps

1. **Security**: JWT Secret in Production verwenden
2. **Database**: SSL für PostgreSQL aktivieren
3. **Monitoring**: Render Logs und Metriken nutzen
4. **Backups**: Render Backup-Features nutzen