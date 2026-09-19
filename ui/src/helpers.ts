// --------------------------------------------------------------
// URL query parameter helpers
// --------------------------------------------------------------

/** Returns the value of a query parameter on the current URL, or undefined when it is absent. */
export function getParameterByName(name: string): string | undefined {
    const value = new URLSearchParams(window.location.search).get(name);
    return value ?? undefined;
}

/**
 * Sets a query parameter on the current URL, replacing any existing value.
 * The page is not reloaded; the change replaces the current history entry so it
 * does not add a back-button step.
 */
export function setParameterByName(name: string, value: string) {
    const url = new URL(window.location.href);
    url.searchParams.set(name, value);
    window.history.replaceState(window.history.state, '', url);
}

/** Removes a query parameter from the current URL. Does nothing when it is not present. */
export function deleteParameterByName(name: string) {
    const url = new URL(window.location.href);
    if (!url.searchParams.has(name)) {
        return;
    }

    url.searchParams.delete(name);
    window.history.replaceState(window.history.state, '', url);
}

// --------------------------------------------------------------
// Login redirects
// --------------------------------------------------------------

export const loginPath = '/login';

const redirectParam = 'redirect';

let redirecting = false;

export function isLoginPage(): boolean {
    return window.location.pathname === loginPath;
}

/**
 * Sends the user to the login page, capturing the current path and query
 * parameters so they can be returned to it afterwards. Called when the API
 * reports an expired or missing session, so it navigates outside of the router
 * to drop any state belonging to the previous session.
 */
export function redirectToLogin() {
    // Several requests can fail at once when a session expires, and the user may
    // already be sitting on the login page.
    if (redirecting || isLoginPage()) {
        return;
    }
    redirecting = true;

    const url = new URL(loginPath, window.location.origin);
    url.searchParams.set(redirectParam, window.location.pathname + window.location.search);
    window.location.assign(url);
}

/**
 * Returns the path captured by redirectToLogin, for the login page to navigate
 * to once the user is authenticated. Falls back to the given path when nothing
 * usable was captured.
 */
export function getPostLoginRedirect(fallback: string = '/'): string {
    const target = getParameterByName(redirectParam);
    if (!target) {
        return fallback;
    }

    // The parameter is attacker-controllable, so only same-site paths are
    // honoured: "//host" and "/\host" are treated as absolute by browsers.
    if (!target.startsWith('/') || target.startsWith('//') || target.startsWith('/\\')) {
        return fallback;
    }

    return target;
}

// --------------------------------------------------------------
// String Helpers
// --------------------------------------------------------------

export function capitalizeWords(str: string): string {
    return str
        .toLowerCase()                          // 1. Lowercase the entire string
        .split(' ')                             // 2. Split it into an array of words
        .map(word => word.charAt(0).toUpperCase() + word.slice(1)) // 3. Capitalize first letter of each
        .join(' ');                             // 4. Join the array back into a string
}