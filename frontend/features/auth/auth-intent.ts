import { Buffer } from "buffer";

import { publicEnv } from "@/lib/env/public";

export const AUTH_INTENT_VALUES = ["public", "seller"] as const;
export type AuthIntent = (typeof AUTH_INTENT_VALUES)[number];

export const AUTH_INTENT_STATE_VERSION = 1;

export type AuthIntentStatePayload = {
  version: typeof AUTH_INTENT_STATE_VERSION;
  intent: AuthIntent;
  returnTo: string;
  nonce: string;
};

type AuthIntentStateContext = {
  intent: AuthIntent;
  returnTo?: string;
  nonce?: string;
};

const DEFAULT_RETURN_PATHS: Record<AuthIntent, string> = {
  public: "/",
  seller: "/dashboard",
};

const GOOGLE_OAUTH_PATH = "/auth/oauth/google";

export function buildAuthIntentState(context: AuthIntentStateContext) {
  const payload: AuthIntentStatePayload = {
    version: AUTH_INTENT_STATE_VERSION,
    intent: context.intent,
    returnTo: normalizeReturnPath(context.intent, context.returnTo),
    nonce: context.nonce ?? generateNonce(),
  };

  return toBase64Url(JSON.stringify(payload));
}

export function parseAuthIntentState(state: string): AuthIntentStatePayload {
  const decoded = fromBase64Url(state);
  const parsed = JSON.parse(decoded);

  if (parsed.version !== AUTH_INTENT_STATE_VERSION) {
    throw new Error("unsupported auth intent state version");
  }

  if (!isAuthIntent(parsed.intent)) {
    throw new Error("invalid auth intent");
  }

  if (typeof parsed.returnTo !== "string" || parsed.returnTo.trim() === "") {
    throw new Error("missing returnTo in auth intent state");
  }

  if (typeof parsed.nonce !== "string" || parsed.nonce.trim() === "") {
    throw new Error("missing nonce in auth intent state");
  }

  return parsed as AuthIntentStatePayload;
}

export function buildGoogleOAuthBeginUrl(options: {
  intent: AuthIntent;
  returnTo?: string;
  baseUrl?: string;
}) {
  const baseUrl = resolveOAuthBaseUrl(options.baseUrl);
  const state = buildAuthIntentState({ intent: options.intent, returnTo: options.returnTo });
  const params = new URLSearchParams({ state });

  return `${baseUrl}?${params.toString()}`;
}

function isAuthIntent(value: unknown): value is AuthIntent {
  return typeof value === "string" && (AUTH_INTENT_VALUES as readonly string[]).includes(value);
}

function normalizeReturnPath(intent: AuthIntent, override?: string) {
  if (override && override.trim() !== "") {
    return override;
  }

  return DEFAULT_RETURN_PATHS[intent];
}

function resolveOAuthBaseUrl(override?: string) {
  if (override) {
    const trimmed = withoutTrailingSlash(override);

    if (trimmed.endsWith(GOOGLE_OAUTH_PATH)) {
      return trimmed;
    }

    return `${trimmed}${GOOGLE_OAUTH_PATH}`;
  }

  const apiBase = withoutTrailingSlash(publicEnv.NEXT_PUBLIC_API_BASE_URL);
  return `${apiBase}${GOOGLE_OAUTH_PATH}`;
}

function withoutTrailingSlash(value: string) {
  return value.replace(/\/+$/, "");
}

function generateNonce() {
  const cryptoApi = typeof globalThis !== "undefined" ? globalThis.crypto : undefined;

  if (cryptoApi && typeof cryptoApi.randomUUID === "function") {
    return cryptoApi.randomUUID();
  }

  return `${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}

function toBase64Url(value: string) {
  const base64 = encodeBase64(value);
  return base64.replace(/=+$/g, "").replace(/\+/g, "-").replace(/\//g, "_");
}

function fromBase64Url(value: string) {
  let base64 = value.replace(/-/g, "+").replace(/_/g, "/");
  const padding = 4 - (base64.length % 4);

  if (padding && padding < 4) {
    base64 += "=".repeat(padding);
  }

  return decodeBase64(base64);
}

function encodeBase64(value: string) {
  if (typeof Buffer !== "undefined") {
    return Buffer.from(value, "utf-8").toString("base64");
  }

  if (typeof btoa === "function" && typeof TextEncoder !== "undefined") {
    const encoder = new TextEncoder();
    const bytes = encoder.encode(value);
    let binary = "";

    for (const byte of bytes) {
      binary += String.fromCharCode(byte);
    }

    return btoa(binary);
  }

  throw new Error("no base64 encoder available");
}

function decodeBase64(value: string) {
  if (typeof Buffer !== "undefined") {
    return Buffer.from(value, "base64").toString("utf-8");
  }

  if (typeof atob === "function" && typeof TextDecoder !== "undefined") {
    const binary = atob(value);
    const bytes = new Uint8Array(binary.length);

    for (let index = 0; index < binary.length; index += 1) {
      bytes[index] = binary.charCodeAt(index);
    }

    const decoder = new TextDecoder();
    return decoder.decode(bytes);
  }

  throw new Error("no base64 decoder available");
}
