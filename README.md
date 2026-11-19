# Sistema Bancario Web

Aplicación bancaria desarrollada en Go. Permite registrar usuarios, iniciar y cerrar sesión, consultar el saldo, realizar retiros y abonos, y consultar un historial exclusivo para cada cuenta.

El proyecto conserva la lógica bancaria y las pruebas de los ápices anteriores y reutiliza los archivos existentes en `headers/`.

## Tecnologías

- Go y biblioteca estándar
- HTML y CSS
- Archivos JSON para persistencia

## Requisitos

- Go instalado
- Windows PowerShell o una terminal compatible

Comprueba la instalación:

```powershell
go version
```

## Ejecución

Clona el repositorio y abre una terminal dentro de su carpeta raíz:

```powershell
git clone https://github.com/ale1593/practica-go-bdd.git
cd practica-go-bdd
go run .
```

Si ya tienes abierta la terminal dentro de la carpeta del proyecto, ejecuta solamente:

```powershell
go run .
```

Abre en el navegador:

http://127.0.0.1:8080

Para detener el servidor pulsa `Ctrl+C`.

## Uso de la aplicación

1. Abre la URL local.
2. Selecciona **Regístrate**.
3. Ingresa nombre del titular, nombre de usuario y contraseña.
4. Inicia sesión con esas credenciales.
5. Consulta el saldo disponible.
6. Realiza un retiro o un abono.
7. Revisa el historial mostrado debajo de las operaciones.
8. Selecciona **Cerrar sesión** para salir.

Al crear una cuenta, el saldo inicial es `$0.00`. Los nombres de usuario se convierten a minúsculas y deben tener entre 3 y 32 caracteres.

## Probar dos usuarios

1. Registra un usuario, por ejemplo `ana`.
2. Inicia sesión y abona o retira efectivo.
3. Cierra la sesión.
4. Registra un segundo usuario, por ejemplo `bruno`.
5. Inicia sesión y verifica que comienza con un saldo independiente.
6. Comprueba que el segundo usuario no muestra las transacciones del primero.

Las cuentas y los saldos se conservan al reiniciar el servidor.

## Pruebas

Ejecuta toda la suite:

```powershell
go test -v ./...
```

Ejecuta también las comprobaciones de concurrencia:

```powershell
go test -race -count=1 ./...
```

Ejecuta el análisis estático:

```powershell
go vet ./...
```

Aplica el formato de Go a los archivos fuente:

```powershell
gofmt -w main.go handlers/*.go headers/*.go
```

## Rutas disponibles

| Ruta | Método | Descripción |
|---|---|---|
| `/` | `GET` | Muestra el login o el panel privado según la sesión |
| `/registro` | `GET`, `POST` | Muestra o procesa el registro |
| `/login` | `POST` | Valida las credenciales y crea la sesión |
| `/logout` | `POST` | Cierra la sesión |
| `/retiro` | `POST` | Retira efectivo de la cuenta autenticada |
| `/abono` | `POST` | Abona efectivo a la cuenta autenticada |
| `/historial` | `GET` | Devuelve exclusivamente el historial JSON del usuario autenticado |
| `/static/` | `GET` | Sirve los recursos CSS |

## Persistencia

- `datos/cuentas.json`: nombre del titular, usuario, hash de contraseña, ID único y saldo.
- `datos/transacciones.json`: historial completo con el `cuenta_id` asociado a las operaciones nuevas.

Las transacciones anteriores sin `cuenta_id` se conservan como registros históricos.

## Estructura

```text
prj_withdrawal/
├── caracteristicas/
├── datos/
│   ├── cuentas.json
│   └── transacciones.json
├── handlers/
│   ├── auth.go
│   ├── web.go
│   └── web_test.go
├── headers/
│   ├── class.go
│   ├── class_test.go
│   ├── test_bdd.go
│   ├── transaccion.go
│   └── transaccion_test.go
├── static/
│   └── css/
│       └── estilos.css
├── templates/
│   ├── index.html
│   ├── login.html
│   └── registro.html
├── go.mod
├── main.go
└── README.md
```
