import { describe, expect, it } from "vitest";

import {
  AUTH_INTENT_STATE_VERSION,
  buildAuthIntentState,
  buildGoogleOAuthBeginUrl,
  parseAuthIntentState,
} from "@/features/auth/auth-intent";

describe("auth intent contract", () => {
  it("defaults to public return path when none provided", () => {
    const state = buildAuthIntentState({ intent: "public", nonce: "nonce-public" });
    const parsed = parseAuthIntentState(state);

    expect(parsed.intent).toBe("public");
    expect(parsed.returnTo).toBe("/");
    expect(parsed.nonce).toBe("nonce-public");
    expect(parsed.version).toBe(AUTH_INTENT_STATE_VERSION);
  });

  it("respects custom return path for seller intent", () => {
    const returnTo = "/seller/onboarding";
    const state = buildAuthIntentState({ intent: "seller", returnTo, nonce: "nonce-seller" });
    const parsed = parseAuthIntentState(state);

    expect(parsed.intent).toBe("seller");
    expect(parsed.returnTo).toBe(returnTo);
  });

  it("builds the oauth begin url with encoded state", () => {
    const customBase = "https://api.example.com/";
    const returnTo = "/dashboard/listings";
    const url = buildGoogleOAuthBeginUrl({ intent: "seller", returnTo, baseUrl: customBase });

    const parsedUrl = new URL(url);

    expect(parsedUrl.origin).toBe("https://api.example.com");
    expect(parsedUrl.pathname).toBe("/auth/oauth/google");

    const stateParam = parsedUrl.searchParams.get("state");
    expect(stateParam).toBeTruthy();

    const parsedState = parseAuthIntentState(stateParam!);
    expect(parsedState.intent).toBe("seller");
    expect(parsedState.returnTo).toBe(returnTo);
  });
});
