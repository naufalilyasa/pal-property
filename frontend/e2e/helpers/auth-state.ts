import { Buffer } from "node:buffer";

import type { AuthIntentStatePayload } from "@/features/auth/auth-intent";

function normalizeBase64Url(value: string) {
  const replaced = value.replace(/-/g, "+").replace(/_/g, "/");
  const padding = (4 - (replaced.length % 4)) % 4;
  return `${replaced}${"=".repeat(padding)}`;
}

export function parseAuthIntentStateFromParam(state: string): AuthIntentStatePayload {
  if (!state) {
    throw new Error("auth intent state is missing");
  }

  const decoded = Buffer.from(normalizeBase64Url(state), "base64").toString("utf-8");
  return JSON.parse(decoded) as AuthIntentStatePayload;
}

export function parseAuthIntentStateFromHref(href: string): AuthIntentStatePayload {
  const stateParam = new URL(href).searchParams.get("state");

  if (!stateParam) {
    throw new Error("auth intent state missing from href");
  }

  return parseAuthIntentStateFromParam(stateParam);
}
