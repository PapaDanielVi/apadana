# apadana

apadana is an sdk that provides useful tools for multi-tenancy for golang applications

## TODOs

- Obj management sdk. Lazy init and Centralized (single instance for all tenants)
- prometheus metrics system
- context management tools
- config management tools
- middlewares to read and ingest tenant id
- tools for opentelemetry (otel) tenant id injection
- timezone per tenant tools
- rabbitmq tools for multitenancy publish and consuming
- kafka consumer and publishing with tenant id
- tenant replacer for replacing tenant id with something else
- nats tools for consuming and publishing with tenant id
- redis v9 tools for reading and writing keys with tenant id
- tenant registry for mysql and postgresql to make the db connections multi-tenant both by column or database or instance.
- wrapper around good logger to log the tenant id too in logs
- tools for http client in order to inject tenant id into the headers and pass to other micro-services.
- burst controller per tenant implementations
