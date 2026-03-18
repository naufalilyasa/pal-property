export type BackendEnvelope<T> = {
  success: boolean;
  message: string;
  data: T;
  trace_id: string;
};

export type NormalizedApiResponse<T> = {
  data: T;
  message: string;
  traceId: string;
};

type ApiErrorOptions = {
  status: number;
  traceId?: string;
  data?: unknown;
};

export class ApiError extends Error {
  status: number;
  traceId?: string;
  data?: unknown;

  constructor(message: string, options: ApiErrorOptions) {
    super(message);
    this.name = "ApiError";
    this.status = options.status;
    this.traceId = options.traceId;
    this.data = options.data;
  }
}

function isBackendEnvelope(value: unknown): value is BackendEnvelope<unknown> {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<BackendEnvelope<unknown>>;

  return (
    typeof candidate.success === "boolean" &&
    typeof candidate.message === "string" &&
    "data" in candidate &&
    typeof candidate.trace_id === "string"
  );
}

export async function normalizeApiResponse<T>(
  response: Response,
): Promise<NormalizedApiResponse<T>> {
  let payload: unknown;

  try {
    payload = await response.json();
  } catch {
    throw new ApiError("Invalid JSON response from API", {
      status: response.status,
    });
  }

  if (!isBackendEnvelope(payload)) {
    throw new ApiError("Invalid API response envelope", {
      status: response.status,
    });
  }

  if (!response.ok || !payload.success) {
    throw new ApiError(payload.message || "API request failed", {
      status: response.status,
      traceId: payload.trace_id,
      data: payload.data,
    });
  }

  return {
    data: payload.data as T,
    message: payload.message,
    traceId: payload.trace_id,
  };
}
