# CLAUDE.md

Terraform provider (Plugin Framework) for Pi-hole API v6: DNS A records, CNAME records and
config settings, as resources and data sources. API, auth flow and endpoints: `SPECS.md`; user
docs incl. the application-password / `webserver.api.app_sudo` limits: `docs/index.md`; layout
and conventions: `CONTRIBUTING.md`.

- `just adlc-verify` is the gate (fmt-check, vet, lint, unit tests; writes nothing). `just --list`
  for the rest; acceptance tests need a live Pi-hole (`README.md`, Testing).
- `docs/` is the published Terraform Registry documentation, not a place for developer notes.
