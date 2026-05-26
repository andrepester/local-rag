# Threat Model v1 (kompakt)

Dieses Dokument beschreibt die Security-Ableitung zur ADR-Entscheidung aus `docs/architecture/RAG-SEARCH-MCP-ADR-2026-04-11-lan-betriebsmodus.md`.

## Scope

- Endpunkte: `/mcp`, `/healthz`
- Betriebsmodus gemaess ADR: `localhost-only` (Default), `LAN-only` (Opt-in)

## Umsetzungsstatus

- `localhost-only` bleibt der Default: Docker published den Host-Port auf Loopback.
- `LAN-only` ist technisch nur mit expliziter Non-Loopback-Publish-Konfiguration und gesetztem `RAG_API_TOKEN` freigegeben.
- Ist `RAG_API_TOKEN` gesetzt, verlangen `/mcp` und `/mcp/` immer `Authorization: Bearer <token>`.
- `/healthz` bleibt bewusst public-minimal und gibt keine Diagnose- oder Konfigurationsdetails preis.

## Bedrohungen (v1)

- Unautorisierte Zugriffe von Clients im LAN
- Fehlkonfiguration durch zu breite Exposition (z. B. WAN/oeffentliche Erreichbarkeit)
- Kompromittierte LAN-Clients mit legitimer Netznaehe

## Verbindliche Controls

- Netzgrenzen: primaer Docker/Host/Firewall, nur freigegebene Netze im LAN-Opt-in
- Authentisierung: Bearer-Token-Pflicht fuer geschuetzte MCP-Pfade im LAN-Opt-in; Non-Loopback-Publish ohne Token ist ungueltige Konfiguration
- CORS: kein permissiver Default
- Discovery: keine automatische Service Discovery in v1

## Test-/Compliance-Checks

- `localhost-default`: mit Default-Config ist Zugriff auf `/mcp` nur lokal erfolgreich.
- `LAN-opt-in`: nur mit expliziter Non-Loopback-Publish-Konfiguration, gesetztem `RAG_API_TOKEN` und dokumentierter Source-Netzgrenze.
- `Auth`: Requests auf `/mcp` oder `/mcp/` ohne gueltiges Bearer-Token werden bei gesetztem Token abgewiesen.
- `Health`: `/healthz` bleibt ohne Token erreichbar und liefert nur ein minimales Health-Signal.
- `Out-of-scope`: keine Exposition ueber WAN/oeffentliche Interfaces, kein VPN/Overlay-Zugriff ohne neues Threat Model.

## Out-of-Scope (v1)

- WAN/Internet-Exposition
- VPN/Overlay-Access ohne separates Threat Model
- Offener Reverse-Proxy/Ingress ins Internet

## Restrisiken

- Fehlkonfiguration auf Host-/Firewall-Ebene
- Kompromittierte Clients im erlaubten LAN-Segment
- Token-Leakage ohne saubere Rotation/Operational Hygiene

## Folgearbeiten in Vikunja

- `P1-004 API Security Baseline (Token-first)`
- `P1-003 MacOS/Linux Harmonisierung fuer Docker-Workflows`
- `P1-009 Observability-Baseline (Metriken, Logs, Health)`
