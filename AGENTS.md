# AGENTS.md

## Project

This is a mature Go library. Preserve API compatibility unless the task
explicitly requires an API change.

## Verification

Run:

    make

before considering a change complete.

## Working principles

- Prefer minimal changes.
- Do not refactor unrelated working code.
- Follow existing code and documentation conventions.
- Treat existing behavior as intentional unless evidence suggests otherwise.
- Check exported Go API documentation when exported behavior changes.
- Keep README and other user documentation consistent with the implementation.
- Do not commit changes unless explicitly requested.
