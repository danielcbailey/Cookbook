import type { User } from './apiTypes';
import { redirectToLogin } from './helpers';
import { invalidateToken } from './shared/authHelpers';

const basePath = '/api/v1/user';

/** Error thrown for any non-2xx response from the user API. */
export class UserAPIError extends Error {
    status: number;

    constructor(status: number, message: string) {
        super(message);
        this.name = 'UserAPIError';
        this.status = status;
    }
}

export type RequestOptions = {
    signal?: AbortSignal;
}

export type LoginResult = {
    token: string;
    /** Unix timestamp, in seconds, at which the session expires. */
    expiry: number;
}

// The API authenticates via the "session" cookie set at login, so every request
// must include credentials. Errors come back as plain text via http.Error.
//
// A 401 means the session is gone or expired and there is nothing the caller can
// do to recover, so the user is sent to log in again — except on the login
// endpoint itself, where it means the credentials were rejected and the user is
// already where they need to be.
async function send(path: string, init: RequestInit, options?: RequestOptions & { redirectOnUnauthorized?: boolean }): Promise<Response> {
    const response = await fetch(basePath + path, {
        ...init,
        credentials: 'include',
        signal: options?.signal,
    });

    if (!response.ok) {
        const body = (await response.text()).trim();

        if (response.status === 401 && options?.redirectOnUnauthorized !== false) {
            invalidateToken();
            redirectToLogin();
        }

        throw new UserAPIError(response.status, body || response.statusText);
    }

    return response;
}

/**
 * Logs the user in and starts a session. The server replies with a "session"
 * cookie that authenticates subsequent API calls, so the returned token only
 * needs to be kept if requests are to be authorized via the Authorization
 * header instead.
 *
 * Set longLived when the user asked to be remembered: the server issues a token
 * with a longer expiry rather than one lasting only the current session.
 *
 * Throws a UserAPIError with status 401 when the credentials are rejected.
 */
export async function login(email: string, password: string, longLived: boolean = false, options?: RequestOptions): Promise<LoginResult> {
    const response = await send('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, long_lived: longLived }),
    }, { ...options, redirectOnUnauthorized: false });

    return await response.json() as LoginResult;
}

/**
 * Returns the signed-in user, including their limits and current usage.
 *
 * The password hash and internal usage-reset bookkeeping are stripped by the
 * server, so the response only carries fields the UI can display.
 */
export async function getCurrentUser(options?: RequestOptions): Promise<User> {
    const response = await send('/me', { method: 'GET' }, options);
    return await response.json() as User;
}
