
## 🚀 Guía de Ejecución Rápida 

El proyecto incluye un entorno de desarrollo en memoria (`devserver`) que permite ejecutar y probar toda la API y el chatbot de manera instantánea, **sin necesidad de levantar bases de datos PostgreSQL ni contenedores Docker** para las pruebas básicas.

### 1. Iniciar el servidor localmente
Clonar el repositorio
Ejecutá el siguiente comando en la terminal:
```bash
make run o make demo


El proyecto incluye un entorno de desarrollo en memoria (`devserver`) que permite ejecutar y probar toda la API y el chatbot de manera instantánea, **sin necesidad de levantar bases de datos PostgreSQL ni contenedores Docker** para las pruebas básicas.

Hay dos formas de probar el chatbot:

| Opción | Cuándo usarla |
| --- | --- |
| **Demo automática (`make demo`)** | Recomendada. Levanta el servidor, ejecuta todos los escenarios y valida las respuestas sin intervención manual. |
| **Pruebas manuales con `curl`** | Para explorar la API a mano, enviando tus propios mensajes. |

### Opción A — Demo automática (recomendada)

Ejecutá un único comando:

```bash
make demo
```
Podés ver el detalle técnico en el [Manual del Chatbot](docs/chatbot.md).
Podés ver el detalle de la arquitectura del [Sistema] (dosc/arquitectura.md)
