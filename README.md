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

## Mensajes de WhatsApp para reservas

En `/admin`, la pestaña **Mensaje de WhatsApp** permite editar y guardar el texto usado por el botón **Escribir** de cada reserva. Se puede combinar texto libre con variables como `{nombre}`, `{fecha}`, `{hora}`, `{personas}`, `{ubicacion}`, `{motivo}`, `{motivo_detalle}`, `{duracion}`, `{estado}`, `{whatsapp}`, `{id}` y `{creada_el}`. Los botones de variables las insertan en la posición del cursor y la vista previa muestra una reserva real cuando existe. Los cambios se guardan en PostgreSQL y se aplican a todas las reservas al pulsar **Guardar mensaje**. El botón abre WhatsApp con el texto preparado; el envío final lo realiza el administrador.

## Imágenes de actividades

El administrador puede cargar JPG, PNG, WebP o AVIF de hasta 20 MB y ajustar un recorte fijo 4:5 con arrastre, zoom y controles accesibles. Al guardar la actividad, el navegador envía únicamente el recorte de 1080 × 1350 píxeles. La API autenticada valida las dimensiones, vuelve a codificar los píxeles y los convierte a AVIF con `avifenc` (incluido en la imagen Docker). Los archivos se sirven desde `/media/activities/` y se conservan en el volumen Docker `uploads`, montado en `/data/uploads`. `UPLOAD_DIR` permite cambiar esa carpeta en ejecuciones fuera de Docker. El volumen debe incluirse en los respaldos junto con PostgreSQL.

## Deploy (Coolify)

Un solo recurso **Docker Compose** apuntando a este repo, rama `main`, archivo `docker-compose.prod.yml`. El archivo `docker-compose.yml` sigue siendo para desarrollo local.

En Coolify, configurar `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD` y `PUBLIC_API_URL`. Usar una contraseña hexadecimal aleatoria para Postgres, ya que se incluye en la URL interna de conexión. El compose de producción construye `DATABASE_URL` con esas variables, exige secretos y conserva Postgres en el volumen `pgdata`.

Dominios de producción: `https://inspirarteperu.com:80` y `https://www.inspirarteperu.com:80` para `web`, y `https://api.inspirarteperu.com:8080` para `backend`. En Coolify el puerto indica el destino interno del proxy; las URLs públicas usan HTTPS estándar. `PUBLIC_API_URL=https://api.inspirarteperu.com` se integra al compilar la web, por lo que cambiarla requiere reconstruirla.

El despliegue automático requiere configurar el webhook de GitHub en Coolify.

### Galería de fotografías

En el dashboard, la pestaña **Galería** permite subir fotografías de hasta 20 MB, editar su descripción (texto alternativo), indicar el orden y ocultarlas o volver a publicarlas. El navegador conserva la proporción y reduce el lado mayor a 1600 px; la API convierte los píxeles a AVIF. Los archivos utilizan el volumen persistente de imágenes existente y los metadatos se guardan en `gallery_images` mediante la migración `003_gallery.sql`. La sección pública muestra una marquesina continua a todo el ancho, con bordes difuminados y pausa al enfocar o pasar el cursor. Con movimiento reducido se convierte en una tira deslizable y aparece cuando existe al menos una foto visible. Ocultar una foto conserva su archivo y registro.
