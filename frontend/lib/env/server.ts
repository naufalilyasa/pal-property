import "server-only";

import { z } from "zod";

const serverEnvSchema = z.object({
  API_BASE_URL: z.string().url().default("http://127.0.0.1:8080"),
  NEXT_PUBLIC_API_BASE_URL: z.string().url().default("http://127.0.0.1:8080"),
});

export const serverEnv = serverEnvSchema.parse({
  API_BASE_URL: process.env.API_BASE_URL,
  NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
});
