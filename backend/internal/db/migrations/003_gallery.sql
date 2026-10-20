CREATE TABLE IF NOT EXISTS gallery_images (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 imagen_url TEXT NOT NULL,
 descripcion TEXT NOT NULL DEFAULT '',
 posicion INT NOT NULL DEFAULT 0,
 activa BOOLEAN NOT NULL DEFAULT TRUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_gallery_order ON gallery_images (posicion, created_at);
