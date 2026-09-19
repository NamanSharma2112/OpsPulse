# 1. Record architecture decisions

- **Status:** accepted
- **Date:** 2026-09-19

## Context

Choices made early in a project — how tenancy works, why there is no ORM,
what authenticates a webhook — get re-litigated every few months unless the
reasoning is written down somewhere durable. Commit messages are too granular
and the README is the wrong audience.

## Decision

Record every architecturally significant decision as a numbered Markdown file
in `docs/decisions/`, in the style Michael Nygard described.

A decision is architecturally significant if reversing it would mean changing
more than one package, a database migration, or an external contract.

Each record states its status (`proposed`, `accepted`, `superseded by NNNN`),
the context that forced the decision, the decision itself, and its
consequences — including the bad ones.

Records are immutable once accepted. A changed mind means a new record that
supersedes the old one, not an edit.

## Consequences

- New contributors can read the decision log instead of asking why.
- Writing a record is a small tax on each significant change.
- The log will contain decisions we later regret. That is the point; a
  superseding record explains what we learned.
