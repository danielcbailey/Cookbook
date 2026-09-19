import type { Ingredient } from './apiTypes';
import { redirectToLogin } from './helpers';
import { invalidateToken } from './shared/authHelpers';

const basePath = '/api/v1/ingredient';

/** Error thrown for any non-2xx response from the ingredient API. */
export class IngredientAPIError extends Error {
    status: number;

    constructor(status: number, message: string) {
        super(message);
        this.name = 'IngredientAPIError';
        this.status = status;
    }
}

export type RequestOptions = {
    signal?: AbortSignal;
}

// The API authenticates via the "session" cookie set at login, so every request
// must include credentials. Errors come back as plain text via http.Error.
async function send(path: string, init: RequestInit, options?: RequestOptions): Promise<Response> {
    const response = await fetch(basePath + path, {
        ...init,
        credentials: 'include',
        signal: options?.signal,
    });

    if (!response.ok) {
        const body = (await response.text()).trim();

        // The session is gone or expired: there is nothing the caller can do to
        // recover, so send the user to log in again.
        if (response.status === 401) {
            invalidateToken();
            redirectToLogin();
        }

        throw new IngredientAPIError(response.status, body || response.statusText);
    }

    return response;
}

async function get<T>(path: string, params: Record<string, string | number>, options?: RequestOptions): Promise<T> {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
        query.set(key, String(value));
    }

    const search = query.toString();
    const response = await send(search ? `${path}?${search}` : path, { method: 'GET' }, options);
    return await response.json() as T;
}

// List endpoints marshal a nil slice as JSON null when the user has no matches.
function asList(ingredients: Ingredient[] | null): Ingredient[] {
    return ingredients ?? [];
}

// --------------------------------------------------------------
// Listing
// --------------------------------------------------------------

/**
 * Lists the ingredients the user owns, ordered by the server. Embeddings are
 * stripped server-side, so the returned ingredients never carry one.
 */
export async function listIngredients(offset: number, limit: number, options?: RequestOptions): Promise<Ingredient[]> {
    return asList(await get<Ingredient[] | null>('/list', { offset, limit }, options));
}
