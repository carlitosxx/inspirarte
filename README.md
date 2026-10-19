# Inspirarte Espacio

Plataforma de reservas para **Inspirarte Espacio** — cafetería, experiencias creativas, talleres y celebraciones.

Monorepo con dos piezas:

```
inspirarte/
├── web/        → Landing + reserva de mesa + actividades del mes + dashboard /admin (Astro, estático)
└── backend/    → API REST (Go + PostgreSQL)
```

## Desarrollo local

```bash
# 1. Variables de entorno
cp .env.example .env

# 2. Levantar todo (Postgres + API + web)
docker compose up --build
```

- Web: http://localhost:3000
- API: http://localhost:8080/health
- Dashboard: http://localhost:3000/admin (login con `ADMIN_EMAIL` / `ADMIN_PASSWORD` del `.env`)

Sin Docker:

```bash
# Backend (requiere Postgres local y DATABASE_URL)
cd backend && go run ./cmd/api

# Web (requiere PUBLIC_API_URL apuntando a la API)
cd web && npm install && npm run dev
```

## Reglas del negocio (datos del cliente)

- **Horario:** lun–vie 5:00 p. m. – 10:00 p. m. · sáb 10:00 a. m. – 10:00 p. m. · dom solo con reserva previa para grupos o celebraciones especiales.
- **Ubicaciones:** mesa en salón, mesa junto a la mampara (máx. 2 personas), barra junto a la ventana.
- **Permanencia:** cafetería 2 h 30 · celebraciones 3 h · talleres según la actividad.
- **Tolerancia:** 15 minutos; pasado ese tiempo, si el cliente no se comunica, la reserva puede liberarse. El aviso aparece en la confirmación de reserva.
- **Actividades del mes:** cupos con descuento atómico (sin overbooking); al agotarse se muestra "Cupos agotados ✨". Se gestionan desde el dashboard.

## Deploy (Coolify)

Un solo recurso **Docker Compose** apuntando a este repo, rama `main`, archivo `docker-compose.yml`. Ver variables en `.env.example`. Cada push a `main` redespliega automáticamente (webhook de GitHub → Coolify).
