import { getOptionalUser } from "@/features/auth/server/current-user";
import type { SellerSession } from "@/lib/session/bootstrap";

export async function getSellerSession(): Promise<SellerSession> {
  const user = await getOptionalUser();

  if (!user) {
    return {
      status: "unauthenticated",
      user: null,
    };
  }

  return {
    status: "authenticated",
    user,
    traceId: "server-session",
  };
}
