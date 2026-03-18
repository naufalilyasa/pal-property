import "server-only";

import { redirect } from "next/navigation";

import { getOptionalUser } from "@/features/auth/server/current-user";

export async function requireUser() {
  const user = await getOptionalUser();

  if (!user) {
    redirect("/login");
  }

  return user;
}
