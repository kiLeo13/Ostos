# API contract

`openapi.yaml` is the source of truth for HTTP behavior. The checked-in `generated/` directory is the shared TypeScript fetch client for mobile and, later, web.

For every contract change:

1. Run `npm run lint`.
2. Review compatibility against the previously released specification. The foundation contract has no predecessor, so its initial review is baseline-only.
3. Run `npm run generate`.
4. Run `npm run check:generated` and include the deterministic generated output with the change.

Server handlers still use dedicated request DTOs and explicit mapping. Generated transport types never become domain entities.
