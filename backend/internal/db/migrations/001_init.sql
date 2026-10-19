CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Reservas de mesa (cafetería, celebraciones, reuniones)
CREATE TABLE IF NOT EXISTS reservations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre        TEXT NOT NULL,
    whatsapp      TEXT NOT NULL,
    fecha         DATE NOT NULL,
    hora          TIME NOT NULL,
    personas      INT NOT NULL CHECK (personas > 0),
    ubicacion     TEXT NOT NULL CHECK (ubicacion IN ('salon', 'mampara', 'barra')),
    motivo        TEXT NOT NULL CHECK (motivo IN ('cafeteria', 'cumpleanos', 'aniversario', 'reunion', 'experiencia', 'otro')),
    motivo_detalle TEXT,
    duracion_min  INT NOT NULL DEFAULT 150,
    estado        TEXT NOT NULL DEFAULT 'pendiente'
                  CHECK (estado IN ('pendiente', 'confirmada', 'cancelada', 'liberada')),
    notas         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reservations_fecha ON reservations (fecha, hora);

-- Actividades del mes (talleres y experiencias con cupos)
CREATE TABLE IF NOT EXISTS activities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo            TEXT NOT NULL,
    descripcion       TEXT,
    imagen_url        TEXT,
    fecha_hora        TIMESTAMPTZ NOT NULL,
    precio            NUMERIC(10, 2) NOT NULL DEFAULT 0,
    cupos_totales     INT NOT NULL CHECK (cupos_totales > 0),
    cupos_disponibles INT NOT NULL CHECK (cupos_disponibles >= 0),
    activa            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_activities_fecha ON activities (fecha_hora);

-- Cupos apartados por clientes en cada actividad
CREATE TABLE IF NOT EXISTS activity_bookings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities (id) ON DELETE CASCADE,
    nombre      TEXT NOT NULL,
    whatsapp    TEXT NOT NULL,
    cupos       INT NOT NULL DEFAULT 1 CHECK (cupos > 0),
    estado      TEXT NOT NULL DEFAULT 'confirmada'
                CHECK (estado IN ('confirmada', 'cancelada')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bookings_activity ON activity_bookings (activity_id);

-- Usuarios administradores (el login valida contra ADMIN_EMAIL/ADMIN_PASSWORD del entorno)
CREATE TABLE IF NOT EXISTS admin_users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
