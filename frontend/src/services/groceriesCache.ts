import type { Grocery } from "@/types/grocery";
import { readJSONArray, writeJSON } from "@/services/storage";

const ALL_KEY = "cached_groceries";
const NEEDED_KEY = "cached_needed";

export function saveAll(items: Grocery[]): void {
    writeJSON(ALL_KEY, items);
}

export function loadAll(): Grocery[] {
    return readJSONArray<Grocery>(ALL_KEY);
}

export function saveNeeded(uuids: string[]): void {
    writeJSON(NEEDED_KEY, uuids);
}

export function loadNeeded(): string[] {
    return readJSONArray<string>(NEEDED_KEY);
}
