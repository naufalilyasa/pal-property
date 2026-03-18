import { publicEnv } from "@/lib/env/public";

export function getBrowserApiBaseUrl(): string {
  return publicEnv.NEXT_PUBLIC_API_BASE_URL.replace(/\/$/, "");
}

export function buildApiUrl(path: string, baseUrl: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${normalizedPath}`;
}
