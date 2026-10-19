# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Personas que quieren visitar Inspirarte Espacio o reservar un cupo en una actividad creativa. La acción principal del sitio es reservar una mesa.

## Product Purpose

Presentar la cafetería y sus experiencias, permitir reservas de mesa y cupos para actividades, y ofrecer contacto directo por WhatsApp.

## Operating Context

El sitio público muestra actividades vigentes y recibe reservas. El panel de administración permite gestionar reservas y actividades. Las confirmaciones de reserva se coordinan por WhatsApp.

## Capabilities and Constraints

- Sitio Astro con API independiente. Los formularios existentes de mesa y actividades y la ruta `/admin` deben seguir funcionando.
- Los horarios, el teléfono, los tiempos de permanencia y la tolerancia de 15 minutos provienen del sitio existente; no se alteran sin confirmación.
- Producción se despliega desde el repositorio mediante Coolify en `inspirarteperu.com`.
- No hay dirección física publicada ni fotografía verificada del local en el repositorio.

## Brand Commitments

El nombre Inspirarte Espacio y la oferta de cafetería, talleres, experiencias creativas y celebraciones. El usuario ha autorizado una nueva dirección visual para el sitio.

## Evidence on Hand

El contenido y los flujos actuales están en `web/src/pages/index.astro` y `web/src/pages/admin/index.astro`. La imagen anterior del espacio apuntaba a un archivo ausente; cualquier imagen nueva de ambiente se presentará como ilustrativa.

## Product Principles

- Hacer evidente el camino para reservar y el de descubrir actividades.
- Comunicar con claridad condiciones reales de la visita.
- Mostrar la experiencia de manera cálida sin inventar datos del local.
- Mantener los formularios accesibles y fáciles de completar en móvil.
