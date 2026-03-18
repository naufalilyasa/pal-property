import {
  bootstrapSellerSession,
  type BootstrapSellerSessionOptions,
  type SellerSession,
} from "@/lib/session/bootstrap";
import { getRequestCookieHeader } from "@/lib/server/cookies";

export async function getSellerSession(
  options: Omit<BootstrapSellerSessionOptions, "cookieHeader"> = {},
): Promise<SellerSession> {
  return bootstrapSellerSession({
    ...options,
    cookieHeader: await getRequestCookieHeader(),
  });
}
