# Levantando os serviços

```bash
podman-compose up --build
```

```bash
curl -X POST http://localhost:8081/cep \
     -H "Content-Type: application/json" \
     -d '{"cep":"05025000"}'
```

# [Zipkin](http://localhost:9411/zipkin/)