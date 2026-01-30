# Copilot Instructions for go-api

## Project Overview
- This is a Go-based web API project using popular components: Gin (web framework), GORM/Ent (ORM), Redis, JWT, session/cookie auth, i18n, and more.
- The architecture is modular, with clear separation:
  - `api/`: Controllers (MVC pattern), handle HTTP requests and coordinate logic
  - `model/`: Database models and DB operations
  - `service/`: Business logic (e.g., user registration, login)
  - `serializer/`: JSON serialization, transforms DB models to API responses
  - `cache/`: Redis caching logic
  - `auth/`: Authentication and authorization helpers
  - `conf/`: Static config, i18n translations in `locales/`
  - `middleware/`: Gin middleware (CORS, JWT, rate limiting, logging, etc.)
  - `util/`: Utility functions
- Swagger docs are in `docs/` and generated via `swag init`.

## Key Workflows
- **Run locally:** `go run main.go` (default port 3000)
- **Build:** `make` (on Linux/macOS)
- **Hot reload:** `air` (if installed)
- **Environment:** Use `.env` for local config (see README for required vars)
- **Swagger:** Update docs with `swag init` after changing API comments

## Patterns & Conventions
- Controllers in `api/` call services in `service/`, which use models in `model/`.
- Use `serializer/` to define API response formats, not direct DB structs.
- Auth can use session (cookie) or JWT; see `middleware/` and `auth/` for details.
- i18n is custom; translation files are in `conf/locales/`.
- Redis is used for caching and session storage (see `cache/`).
- Use Gin middleware for cross-cutting concerns (CORS, logging, rate limiting).
- ORM: Both GORM and Ent are present; check usage in `model/` and `ent/`.

## Integration Points
- MySQL and Redis required; DSNs set via env or `.env`.
- Swagger UI available after running, see `docs/` for config.
- JetBrains IDEs supported (see README for badge).

## Examples
- User registration: see `api/user.go`, `service/user_register_service.go`, `model/user.go`.
- Middleware: see `middleware/` for CORS, JWT, rate limiting, logging.
- i18n: see `conf/i18n.go` and `conf/locales/zh-cn.yaml`.

## Tips for AI Agents
- Follow the controller → service → model pattern for new features.
- Reuse existing middleware and utilities where possible.
- Keep business logic out of controllers; place in `service/`.
- Update Swagger docs and run `swag init` after API changes.
- Reference `.env` and README for required environment variables.

---
If any section is unclear or missing, please provide feedback for improvement.
