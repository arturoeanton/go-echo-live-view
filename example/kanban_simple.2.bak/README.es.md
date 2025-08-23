# Tablero Kanban Simple

Un tablero Kanban colaborativo en tiempo real construido con el framework Go Echo LiveView. Incluye funcionalidad de arrastrar y soltar, almacenamiento persistente y sincronización automática entre múltiples usuarios.

## Características

- **Colaboración en Tiempo Real**: Los cambios se sincronizan instantáneamente entre todos los usuarios conectados vía WebSocket
- **Arrastrar y Soltar**: 
  - Mover tarjetas entre columnas
  - Reordenar columnas arrastrando sus encabezados
- **Gestión de Tarjetas**:
  - Crear, editar y eliminar tarjetas
  - Asignar niveles de prioridad (Baja, Media, Alta, Urgente)
  - Agregar puntos de historia para estimación
  - Escribir descripciones
- **Gestión de Columnas**:
  - Crear y editar columnas
  - Colores personalizados para organización visual
  - Totales automáticos de tarjetas y puntos
- **Almacenamiento Persistente**: Todos los datos se guardan en `kanban_board.json`
- **Interfaz Limpia**: Diseño moderno y responsivo con animaciones suaves

## Instalación

### Prerrequisitos

- Go 1.19 o superior
- El framework go-echo-live-view (proyecto padre)

### Configuración

1. Navegar al directorio del ejemplo:
```bash
cd example/kanban_simple
```

2. Ejecutar la aplicación:
```bash
go run .
```

3. Abrir el navegador y navegar a:
```
http://localhost:8080
```

## Uso

### Gestión de Tarjetas

- **Agregar Tarjeta**: Hacer clic en el botón "+ Add Card" en cualquier columna
- **Editar Tarjeta**: Hacer clic en cualquier tarjeta para abrir el modal de edición
- **Mover Tarjetas**: Arrastrar y soltar tarjetas entre columnas
- **Prioridad de Tarjeta**: Establecer niveles de urgencia con insignias de colores
- **Puntos de Historia**: Asignar estimación de esfuerzo (0-100 puntos)

### Propiedades de las Tarjetas

- **Título**: El título principal de la tarjeta (requerido)
- **Descripción**: Detalles adicionales sobre la tarea
- **Prioridad**: Establecer nivel de urgencia (Baja, Media, Alta, Urgente)
- **Puntos**: Puntos de historia para estimación de esfuerzo (0-100)
- **Columna**: A qué columna pertenece la tarjeta

### Gestión de Columnas

- **Agregar Columna**: Hacer clic en el botón "+ Add Column"
- **Editar Columna**: Doble clic en cualquier encabezado de columna
- **Reordenar Columnas**: Arrastrar encabezados de columna para reorganizar (intercambiar posiciones)
- **Colores de Columna**: Personalizar colores para mejor organización visual

### Indicadores Visuales

- **Insignia de Puntos**: Insignia azul mostrando puntos de historia (esquina inferior derecha de las tarjetas)
- **Insignias de Prioridad**: Indicadores de prioridad con código de colores
  - Gris: Prioridad baja
  - Naranja: Prioridad media
  - Rojo: Prioridad alta
  - Púrpura: Urgente
- **Estadísticas de Columna**: El encabezado muestra total de tarjetas y puntos

## Almacenamiento de Datos

El tablero guarda automáticamente todos los cambios en `kanban_board.json`. Este archivo contiene:

- Definiciones de columnas (id, título, color, orden)
- Datos de tarjetas (id, título, descripción, columna, prioridad, puntos, marcas de tiempo)

### Estructura de Datos de Ejemplo

```json
{
  "columns": [
    {
      "id": "todo",
      "title": "To Do",
      "color": "#e3e8ef",
      "order": 0
    },
    {
      "id": "doing",
      "title": "In Progress",
      "color": "#ffd4a3",
      "order": 1
    },
    {
      "id": "done",
      "title": "Done",
      "color": "#a3e4d7",
      "order": 2
    }
  ],
  "cards": [
    {
      "id": "card_1755897826",
      "title": "Tarea de Ejemplo",
      "description": "Descripción de la tarea aquí",
      "column_id": "todo",
      "priority": "medium",
      "points": 5,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## Detalles Técnicos

### Arquitectura

- **Backend**: Go con framework web Echo
- **Tiempo Real**: Conexiones WebSocket vía LiveView
- **Frontend**: HTML renderizado del lado del servidor con actualizaciones DOM en tiempo real
- **Almacenamiento**: Persistencia basada en archivos JSON con protección mutex
- **Sincronización**: Gestión de estado global entre todos los clientes conectados

### Componentes Clave

1. **SimpleKanbanModal**: Componente principal que maneja toda la lógica del tablero
2. **KanbanBoardData**: Estructura de datos para columnas y tarjetas
3. **Gestión de Estado Global**: Estado sincronizado entre todos los clientes conectados
4. **Manejadores de Eventos**: Procesa interacciones del usuario (arrastrar, soltar, clic, etc.)

### Eventos WebSocket

- `MoveCard`: Se activa cuando las tarjetas se arrastran entre columnas
- `EditCard`: Abre el modal de edición de tarjetas
- `AddCard`: Crea una nueva tarjeta en una columna
- `EditColumn`: Abre el modal de edición de columna
- `AddColumn`: Crea una nueva columna
- `ReorderColumns`: Maneja el arrastrar y soltar de columnas (intercambio)
- `SaveModal`: Persiste los cambios del formulario del modal
- `UpdateFormField`: Actualizaciones de campos de formulario en tiempo real
- `CloseModal`: Cierra el modal activo

## Desarrollo

### Estructura de Archivos

```
kanban_simple/
├── main.go                 # Punto de entrada de la aplicación
├── simple_kanban_modal.go  # Componente principal Kanban
├── kanban_board.json      # Almacenamiento de datos persistente
├── README.md              # Documentación en inglés
└── README.es.md           # Este archivo (español)
```

### Estructura del Código

El componente principal (`SimpleKanbanModal`) incluye:

- **Estructuras de Datos**: `KanbanColumn`, `KanbanCard`, `KanbanBoardData`
- **Gestión de Estado**: Estado global protegido con mutex
- **Manejadores de Eventos**: Todos los manejadores de interacción del usuario
- **Plantilla**: HTML/CSS/JS completo en `GetTemplate()`
- **Métodos Auxiliares**: `GetCardsForColumn()`, `GetCardCount()`, `GetColumnPoints()`, etc.

### Extender la Aplicación

Para agregar nuevas características:

1. Agregar manejadores de eventos al mapa `Events` en `Start()`
2. Implementar el método manejador en `SimpleKanbanModal`
3. Actualizar la plantilla para incluir elementos de UI
4. Agregar campos necesarios a las estructuras de datos
5. Actualizar la persistencia JSON si es necesario

### Ejecutar en Desarrollo

Para recarga automática durante el desarrollo:

```bash
# Instalar gomon si no está instalado
go install github.com/c9s/gomon@latest

# Ejecutar con recarga automática
gomon
```

## Compatibilidad de Navegadores

- Chrome/Edge (recomendado)
- Firefox
- Safari
- Cualquier navegador moderno con soporte WebSocket

## Características Clave Explicadas

### Reordenamiento de Columnas
Las columnas pueden reordenarse arrastrando sus encabezados. El sistema usa un mecanismo simple de intercambio - cuando sueltas una columna sobre otra, intercambian posiciones.

### Puntos de Historia
Cada tarjeta puede tener puntos de historia (0-100) para estimación de esfuerzo. El total de puntos por columna se muestra en el encabezado de la columna.

### Sincronización en Tiempo Real
Todos los cambios se transmiten inmediatamente a todos los usuarios conectados. El sistema usa un gestor de estado global con protección mutex para garantizar la consistencia de datos.

### Almacenamiento Persistente
Cada cambio activa un guardado automático en `kanban_board.json`. El sistema carga este archivo al iniciar, asegurando la persistencia de datos entre reinicios del servidor.

## Limitaciones Conocidas

- Almacenamiento basado en archivos (no adecuado para uso en producción con alto tráfico)
- Sin autenticación/autorización de usuarios
- Sin archivado/eliminación de tarjetas (las tarjetas permanecen en el sistema)
- Instancia de tablero único (sin soporte multi-tablero)
- Sin funcionalidad de deshacer/rehacer
- Sin capacidades de búsqueda o filtrado

## Consideraciones de Rendimiento

- Adecuado para equipos pequeños a medianos (hasta ~50 usuarios concurrentes)
- El archivo JSON puede manejar miles de tarjetas eficientemente
- Las conexiones WebSocket son ligeras y responsivas
- La protección mutex garantiza seguridad de hilos pero puede impactar el rendimiento bajo carga pesada

## Contribuir

¡Siéntete libre de enviar problemas y solicitudes de mejora! Algunas ideas para contribuciones:

- Agregar autenticación de usuarios
- Implementar archivado/eliminación de tarjetas
- Agregar funcionalidad de búsqueda y filtro
- Crear plantillas de tablero
- Agregar soporte para adjuntos en tarjetas
- Implementar registro de actividad

## Licencia

Este ejemplo es parte del proyecto go-echo-live-view y sigue los mismos términos de licencia.