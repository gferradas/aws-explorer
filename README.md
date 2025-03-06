# aws-explorer

## Descripción

El paquete principal proporciona una aplicación TUI (Interfaz de Usuario de Texto) para explora stacks de AWS CloudFormation. Permite a los usuarios escanear y lista stacks de CloudFormation, cargar perfiles de configuración de AWS y ejecutar acciones como desplega stacks o crear Change-Sets.

La aplicación utiliza la biblioteca [tview](https://github.com/rivo/tview) para crear la TUI e interactúa con AWS CLI para realizar operaciones.

La aplicación soporta interacciones con el ratón y actualiza dinámicamente la interfaz de usuario según las entradas y acciones del usuario.

## Requirements

Para ejecutar aws-explorer, necesitarás:

- AWS CLI instalado y configurado.
- Credenciales de AWS configuradas.
- Un terminal compatible con TUI.
## Instalación

Puedes ejecutar aws-explorer de las siguientes maneras:

### Compilar desde el código fuente

1. Clona el repositorio:
    ```sh
    git clone https://github.com/gferradas/aws-explorer.git
    cd aws-explorer
    ```

2. Instala las dependencias necesarias:
    ```sh
    go mod tidy
    ```

3. Ejecuta la aplicación:
    ```sh
    go run .
    ```

### Descargar paquetes precompilados

1. Ve a la [sección de releases](https://github.com/gferradas/aws-explorer/releases) del repositorio.

2. Descarga el paquete compatible con tu sistema operativo (Linux o macOS).

3. Extrae el archivo descargado y ejecuta la aplicación:
    ```sh
    ./aws-explorer
    ```

