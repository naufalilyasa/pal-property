import { AuthIntent } from "@/features/auth/auth-intent";

export const PUBLIC_HOME_PATH = "/";
export const SELLER_DASHBOARD_PATH = "/dashboard";
export const SELLER_ONBOARDING_PATH = "/seller/onboarding";

export type SellerCapabilityInfo = {
  canAccessDashboard?: boolean;
  requiresOnboarding?: boolean;
};

export function resolveSellerDestination(capabilities?: SellerCapabilityInfo) {
  if (capabilities?.requiresOnboarding) {
    return SELLER_ONBOARDING_PATH;
  }

  if (capabilities?.canAccessDashboard === false) {
    return SELLER_ONBOARDING_PATH;
  }

  return SELLER_DASHBOARD_PATH;
}

export function resolveAuthIntentDestination(
  intent: AuthIntent,
  capabilities?: SellerCapabilityInfo,
) {
  if (intent === "seller") {
    return resolveSellerDestination(capabilities);
  }

  return PUBLIC_HOME_PATH;
}

export function getLoginPathForIntent(intent: AuthIntent) {
  return intent === "seller" ? "/seller/login" : "/login";
}
