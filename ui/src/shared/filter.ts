import { useMemo } from "react";

const DIACRITICS = /\p{Diacritic}/gu;

// Lowercase and strip accents so "Crème" and "creme" compare equal.
function normalize(value: string): string {
    return value.normalize('NFD').replace(DIACRITICS, '').toLowerCase();
}

// Split once per search rather than once per option.
function toKeywords(inputValue: string): string[] {
    return normalize(inputValue).split(/\s+/).filter(Boolean);
}

// Matches when every keyword appears somewhere in the text, in any order.
// No keywords matches everything.
function contains(textValue: string, keywords: string[]): boolean {
    const haystack = normalize(textValue);
    return keywords.every((keyword) => haystack.includes(keyword));
}

export function exactMatch(a: string, b: string): boolean {
    const aNorm = normalize(a);
    const bNorm = normalize(b);

    return aNorm.trim() === bNorm.trim();
}

export interface SearchHaystackKeyItem {
    key: string;
};

export interface SearchHaystackNameItem {
    name: string;
};

export function useSearchFilter<T extends SearchHaystackKeyItem>(search: string, haystack: T[], maxMatches: number): T[] {
    return useMemo(() => {
        const keywords = toKeywords(search);
        const found: T[] = [];
        for (const option of haystack) {
            if (!contains(option.key, keywords)) continue;
            found.push(option);
            if (found.length === maxMatches) break;
        }
        return found;
    }, [search, haystack, maxMatches]);
}

export function useSearchFilterName<T extends SearchHaystackNameItem>(search: string, haystack: T[], maxMatches: number): T[] {
    return useMemo(() => {
        const keywords = toKeywords(search);
        const found: T[] = [];
        for (const option of haystack) {
            if (!contains(option.name, keywords)) continue;
            found.push(option);
            if (found.length === maxMatches) break;
        }
        return found;
    }, [search, haystack, maxMatches]);
}