import type { Recipe, RecipeListItem } from './apiTypes';
import { redirectToLogin } from './helpers';
import { invalidateToken } from './shared/authHelpers';

const basePath = '/api/v1/recipe';

/** Error thrown for any non-2xx response from the recipe API. */
export class RecipeAPIError extends Error {
    status: number;

    constructor(status: number, message: string) {
        super(message);
        this.name = 'RecipeAPIError';
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

        throw new RecipeAPIError(response.status, body || response.statusText);
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

async function postJSON<T>(path: string, payload: unknown, options?: RequestOptions): Promise<T> {
    const response = await send(path, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    }, options);

    return await response.json() as T;
}

// List endpoints marshal a nil slice as JSON null when the user has no matches.
function asList(listings: RecipeListItem[] | null): RecipeListItem[] {
    return listings ?? [];
}

function asOptions(values: string[] | null): string[] {
    return values ?? [];
}

// --------------------------------------------------------------
// Import / save / delete
// --------------------------------------------------------------

/**
 * Extracts a recipe from a web page. The returned recipe is not persisted yet —
 * pass it to saveRecipe once the user confirms it.
 */
export async function importRecipeFromWeb(url: string, options?: RequestOptions): Promise<Recipe> {
    return await postJSON<Recipe>('/importweb', { url }, options);
}

/**
 * Extracts a recipe from photos of it. The combined upload must stay under 20 MiB.
 * As with importRecipeFromWeb, the result is not persisted until it is saved.
 */
export async function importRecipeFromPhotos(photos: File[], options?: RequestOptions): Promise<Recipe> {
    const form = new FormData();
    for (const photo of photos) {
        form.append('photos', photo);
    }

    const response = await send('/importphotos', { method: 'POST', body: form }, options);
    return await response.json() as Recipe;
}

/** Creates or updates a recipe, returning its ID. */
export async function saveRecipe(recipe: Recipe, options?: RequestOptions): Promise<number> {
    const result = await postJSON<{ id: number }>('/save', recipe, options);
    return result.id;
}

export async function deleteRecipe(id: number, options?: RequestOptions): Promise<void> {
    await send('/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id }),
    }, options);
}

// --------------------------------------------------------------
// Listing
// --------------------------------------------------------------

export async function listRecipes(offset: number, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/list', { offset, limit }, options));
}

export async function listRecipesByCategory(category: string, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/listbycategory', { category, limit }, options));
}

export async function listRecipesByProtein(protein: string, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/listbyprotein', { protein, limit }, options));
}

export async function listRecipesByMeal(mealtime: string, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/listbymeal', { mealtime, limit }, options));
}

export async function listRecipesByTag(tag: string, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/listbytag', { tag, limit }, options));
}

/** Searches recipes semantically when the server has embeddings enabled, by title otherwise. */
export async function searchRecipes(query: string, limit: number, options?: RequestOptions): Promise<RecipeListItem[]> {
    return asList(await get<RecipeListItem[] | null>('/search', { query, limit }, options));
}

// --------------------------------------------------------------
// Filter options
// --------------------------------------------------------------

export async function listCategories(options?: RequestOptions): Promise<string[]> {
    return asOptions(await get<string[] | null>('/categories', {}, options));
}

export async function listProteins(options?: RequestOptions): Promise<string[]> {
    return asOptions(await get<string[] | null>('/proteins', {}, options));
}

export async function listMealtimes(options?: RequestOptions): Promise<string[]> {
    return asOptions(await get<string[] | null>('/mealtimes', {}, options));
}

/** Lists every tag name the user owns. A tag may no longer be on any recipe. */
export async function listTags(options?: RequestOptions): Promise<string[]> {
    return asOptions(await get<string[] | null>('/tags', {}, options));
}

// --------------------------------------------------------------
// CRUD operations
// --------------------------------------------------------------

export async function getRecipe(id: number, options?: RequestOptions): Promise<Recipe> {
    return await get<Recipe>('/get/' + id, {}, options);
}
