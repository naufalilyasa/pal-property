import "server-only";

import { redirect } from "next/navigation";

import { AuthIntent } from "@/features/auth/auth-intent";
import { getLoginPathForIntent } from "@/features/auth/auth-destination";
import { getOptionalUser } from "@/features/auth/server/current-user";

type RequireUserOptions = {
  intent?: AuthIntent;
  returnTo?: string;
};

export async function requireUser(options: RequireUserOptions = {}) {
  const { intent = "public", returnTo } = options;
  const user = await getOptionalUser();

  if (!user) {
    const loginPath = getLoginPathForIntent(intent);
    const query = new URLSearchParams();

    if (returnTo) {
      query.set("returnTo", returnTo);
    }

    const redirectTarget = query.toString() ? `${loginPath}?${query}` : loginPath;

    redirect(redirectTarget);
  }

  return user;
}
