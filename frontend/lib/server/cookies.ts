import { cookies } from "next/headers";

export async function getRequestCookieHeader(): Promise<string | undefined> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore
    .getAll()
    .map(({ name, value }) => `${name}=${value}`)
    .join("; ");

  return cookieHeader || undefined;
}
