# Darts Training App - API Dokumentation

## Base URL
`http://localhost:8080/api`

## Authentication
Alle geschützten Endpunkte erfordern einen `Authorization: Bearer <token>` Header.

## Endpunkte

### Authentifizierung

#### POST /auth/login
Login mit Auth0 Code-Exchange.
```json
{
  "code": "authorization_code",
  "state": "random_state"
}
```

#### GET /auth/me
Ruft die aktuellen Benutzerinformationen ab.

### Mannschaften

#### GET /teams
Alle Mannschaften abrufen.

#### POST /teams
Neue Mannschaft erstellen.
```json
{
  "name": "Dart Legends",
  "logo_url": "https://example.com/logo.png"
}
```

#### GET /teams/{id}
Mannschaft nach ID abrufen.

#### PUT /teams/{id}
Mannschaft aktualisieren.
```json
{
  "name": "Updated Team Name",
  "logo_url": "https://example.com/new-logo.png"
}
```

#### DELETE /teams/{id}
Mannschaft löschen.

#### GET /teams/{id}/players
Alle Spieler einer Mannschaft abrufen.

### Spieler

#### GET /players
Alle Spieler abrufen (optional Filter `?team_id={id}`).

#### POST /players
Neuen Spieler erstellen.
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "nickname": "Johnny",
  "is_captain": false,
  "team_id": "uuid-team-id"
}
```

#### GET /players/{id}
Spieler nach ID abrufen.

#### PUT /players/{id}
Spieler aktualisieren.

#### DELETE /players/{id}
Spieler löschen.

#### GET /players/team/{teamId}
Spieler einer Mannschaft abrufen.

#### GET /players/me
Aktuellen Benutzer-Profil abrufen.

#### POST /players/me
Profil für aktuellen Benutzer erstellen.

### Trainingsspiele

#### GET /training-sessions
Alle Trainingsspiele abrufen.

#### POST /training-sessions
Neues Training erstellen.
```json
{
  "name": "Weekly Training",
  "description": "Regular practice session",
  "training_date": "2024-01-15T19:00:00Z",
  "cost_per_player": 5.00
}
```

#### GET /training-sessions/{id}
Training nach ID abrufen.

#### PUT /training-sessions/{id}
Training aktualisieren.
```json
{
  "name": "Updated Training Name",
  "status": "active"
}
```

#### DELETE /training-sessions/{id}
Training löschen.

#### POST /training-sessions/{id}/start
Training starten.

#### POST /training-sessions/{id}/finish
Training beenden.

#### GET /training-sessions/{id}/costs
Kostenberechnung für Training abrufen.

#### POST /training-sessions/{id}/players
Gastspieler hinzufügen.
```json
{
  "guest_name": "Guest Player"
}
```

#### DELETE /training-sessions/players/{playerId}
Spieler vom Training entfernen (nur Gäste).

### Spiele

#### GET /games/modes
Alle verfügbaren Spielmodi abrufen.

#### GET /games/training/{sessionId}
Spiele eines Trainings abrufen.

#### POST /games/training/{sessionId}
Neues Spiel erstellen.
```json
{
  "game_mode_id": "uuid-game-mode-id",
  "player1_id": "uuid-player-1-id",
  "player2_id": "uuid-player-2-id",
  "guest1_name": "Guest 1",
  "guest2_name": "Guest 2"
}
```

#### POST /games/training/{sessionId}/generate
Spiele automatisch generieren (Round-Robin).
```json
{
  "game_mode_id": "uuid-game-mode-id"
}
```

#### PUT /games/{id}
Spiel aktualisieren.
```json
{
  "player1_score": 301,
  "player2_score": 0,
  "status": "completed",
  "winner": "player1"
}
```

#### DELETE /games/{id}
Spiel löschen.

## Status-Codes

- `200 OK` - Erfolgreiche Anfrage
- `201 Created` - Ressource erfolgreich erstellt
- `400 Bad Request` - Ungültige Anfragedaten
- `401 Unauthorized` - Fehlende oder ungültige Authentifizierung
- `404 Not Found` - Ressource nicht gefunden
- `409 Conflict` - Ressource existiert bereits oder Konflikt
- `500 Internal Server Error` - Serverfehler

## Beispiel-Responses

### Team Response
```json
{
  "id": "uuid",
  "name": "Dart Legends",
  "logo_url": "https://example.com/logo.png",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "player_count": 5
}
```

### Training Session Response
```json
{
  "id": "uuid",
  "name": "Weekly Training",
  "description": "Regular practice session",
  "training_date": "2024-01-15T19:00:00Z",
  "cost_per_player": 5.00,
  "status": "active",
  "created_by": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "player_count": 8,
  "game_count": 12,
  "creator_name": "John Doe",
  "training_players": [...],
  "games": [...]
}
```

### Training Costs Response
```json
{
  "training_session_id": "uuid",
  "player_costs": [
    {
      "player_id": "uuid",
      "player_name": "John Doe",
      "is_guest": false,
      "total_cost": 5.00,
      "games_played": 3
    }
  ],
  "total_collected": 40.00
}
```