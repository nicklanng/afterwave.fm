# API specification

This directory contains the **OpenAPI 3** specification for the Afterwave.fm API. The spec is **documentation-only**: the server is implemented in Go and this file is not used to generate any code. When the API changes, update `openapi.yaml` by hand to keep docs in sync.

## Viewing the docs

You can view the spec with any OpenAPI-compatible viewer without generating or running server code:

- **Swagger UI** (interactive): Use the [Swagger Editor](https://editor.swagger.io/) and paste the contents of `openapi.yaml`, or run Swagger UI locally and point it at the file.
- **Redoc**: [Redoc](https://redocly.github.io/redoc/) or `npx @redocly/cli preview-docs api/openapi.yaml` for a read-only doc.
- **VS Code**: Install an extension like “OpenAPI (Swagger) Editor” to preview and validate the spec.

## Validating the spec

```bash
# Optional: validate with Redocly CLI
npx @redocly/cli lint api/openapi.yaml
```

## Not used for

- **Server code generation**: The Go handlers and router are the source of truth; this spec is not used to generate server stubs or routes.
- **Client code generation**: You may use the spec to generate client SDKs if you want; the repo does not do so by default.
