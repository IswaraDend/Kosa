const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
  status: number;
  body: unknown;

  constructor(message: string, status: number, body: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

async function request<T>(
  path: string,
  options: { method?: string; body?: unknown } = {},
): Promise<T> {
  const token = localStorage.getItem('token');
  const res = await fetch(`${API_URL}${path}`, {
    method: options.method ?? 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });

  if (!res.ok) {
    const errorBody = await res.json().catch(() => null);
    throw new ApiError(errorBody?.error ?? res.statusText, res.status, errorBody);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json();

  // Go serializes a nil slice as `null`, not `[]` — list endpoints follow the
  // {data, total} convention, so normalize a null `data` field to an empty
  // array here once instead of null-guarding every page that consumes it.
  if (body && typeof body === 'object' && 'data' in body && body.data === null) {
    body.data = [];
  }

  return body as T;
}

export function buildQuery(params: Record<string, string | number | undefined | null>): string {
  const entries = Object.entries(params).filter(([, v]) => v !== undefined && v !== null && v !== '');
  if (entries.length === 0) return '';
  const search = new URLSearchParams(entries.map(([k, v]) => [k, String(v)]));
  return `?${search.toString()}`;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) => request<T>(path, { method: 'POST', body }),
  put: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PATCH', body }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};
