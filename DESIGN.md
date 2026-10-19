---
name: Inspirarte Espacio
description: Un taller editorial para reservar una mesa y descubrir experiencias creativas.
colors:
  ink: "#33243b"
  plum: "#49305a"
  paper: "#f7f1e9"
  white-paper: "#fffaf4"
  lavender: "#cbb9d9"
  butter: "#f4d58a"
  blush: "#efc7d0"
  sage: "#d9e7e5"
typography:
  display:
    fontFamily: "Fraunces, Georgia, serif"
    fontSize: "clamp(3.5rem, 6vw, 6rem)"
    fontWeight: 600
    lineHeight: 0.97
    letterSpacing: "-.025em"
  body:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: "normal"
rounded:
  control: "9px"
  panel: "16px"
  pill: "100px"
spacing:
  tight: "8px"
  medium: "24px"
  wide: "48px"
components:
  button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.white-paper}"
    rounded: "{rounded.pill}"
    padding: "12px 24px"
  button-hero:
    backgroundColor: "{colors.butter}"
    textColor: "{colors.ink}"
    rounded: "{rounded.pill}"
    padding: "12px 24px"
  input:
    backgroundColor: "#ffffff"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "11px 14px"
---

## Overview

**Creative North Star: “Taller editorial”.** La web se siente como una invitación a sentarse a crear: tipografía expresiva, papel cálido, fotografía táctil y una ruta directa hacia la reserva. El ciruela sostiene los momentos de alta atención; crema y lavanda dejan respirar formularios y contenido.

## Colors

El texto principal usa tinta ciruela sobre papel claro. Amarillo manteca identifica la acción principal sobre fondo oscuro. Lavanda y rosa aportan matices; el verde azulado se reserva para contacto y acentos. No se usa texto de bajo contraste sobre pasteles.

## Typography

Fraunces da voz a titulares y frases cortas. DM Sans mantiene legibles navegación, formularios, horarios y estados. Las cursivas se usan para una sola frase expresiva por bloque, sin convertir todo el contenido en decoración.

## Layout

La portada alterna superficies amplias con composiciones de dos columnas; en móvil todo se apila y la reserva queda en una sola columna. El contenido alcanza un máximo de 1200px y usa márgenes laterales de 24px en escritorio y 16px en móvil.

## Elevation & Depth

El relieve es suave y puntual: la hoja del formulario y el apilado de papeles tienen sombra difusa. Las tarjetas informativas usan borde fino o color de superficie, sin mezclarlo con sombras fuertes.

## Shapes

Los controles principales son cápsulas. Campos y paneles usan esquinas discretas de 9–16px. La fotografía del hero tiene una sola esquina curvada en gran escala, como una página recortada.

## Components

Los botones tienen estados de hover, foco y deshabilitado. Los campos conservan etiquetas visibles y contorno de foco. Las actividades son tarjetas horizontales desplazables con estado vacío propio. El menú móvil es un disclosure nativo.

## Do's and Don'ts

- Mantener el llamado a reservar visible en la primera pantalla y repetirlo cerca de la información útil.
- Usar imágenes de ambiente generadas solo como ilustrativas; nunca hacerlas pasar por fotos verificadas del local.
- Evitar emojis como iconografía y bloques de tarjetas idénticas para estructurar la página.
