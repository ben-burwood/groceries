import type { PendingCreate, PendingNeeded } from "@/types/grocery";
import { readJSONArray, writeJSON } from "@/services/storage";
import { apiFetch } from "@/services/api";

const CREATES_KEY = "pending_creates";
const NEEDED_KEY = "pending_needed";

export type FlushOutcome = {
    synced: number;
    rejected: { kind: "create" | "needed"; uuid: string; status: number }[];
    remaining: number;
};

export type QueueSnapshot = {
    creates: PendingCreate[];
    needed: PendingNeeded[];
};

// --- creates ---
export function listCreates(): PendingCreate[] {
    return readJSONArray<PendingCreate>(CREATES_KEY);
}

export function enqueueCreate(c: PendingCreate): void {
    const items = listCreates();
    if (items.some((i) => i.uuid === c.uuid)) return;
    items.push(c);
    writeJSON(CREATES_KEY, items);
}

// --- needed (collapse on UUID — last write wins) ---
export function listNeeded(): PendingNeeded[] {
    return readJSONArray<PendingNeeded>(NEEDED_KEY);
}

export function enqueueNeeded(op: PendingNeeded): void {
    const items = listNeeded().filter((i) => i.uuid !== op.uuid);
    items.push(op);
    writeJSON(NEEDED_KEY, items);
}

export function snapshot(): QueueSnapshot {
    return { creates: listCreates(), needed: listNeeded() };
}

export function totalPending(snap: QueueSnapshot = snapshot()): number {
    return snap.creates.length + snap.needed.length;
}

// Layers queued setNeeded ops on top of the server's needed-UUID list. Pass a
// pre-fetched snapshot to avoid re-reading localStorage on hot paths.
export function applyNeededOverlay(serverNeeded: string[], queued: PendingNeeded[] = listNeeded()): Set<string> {
    const overlay = new Set(serverNeeded);
    for (const op of queued) {
        if (op.needed) overlay.add(op.uuid);
        else overlay.delete(op.uuid);
    }
    return overlay;
}

async function drainQueue<T extends { uuid: string }>(
    key: string,
    items: T[],
    kind: "create" | "needed",
    request: (item: T) => Promise<Response>,
    rejected: FlushOutcome["rejected"],
): Promise<{ syncedCount: number; remainingCount: number; halted: boolean }> {
    let syncedCount = 0;
    const remaining: T[] = [];
    let halted = false;

    for (const item of items) {
        if (halted) {
            remaining.push(item);
            continue;
        }
        let res: Response;
        try {
            res = await request(item);
        } catch {
            // Network error — keep this item (and the rest) queued.
            halted = true;
            remaining.push(item);
            continue;
        }
        if (res.ok) {
            syncedCount++;
            continue;
        }
        if (res.status >= 400 && res.status < 500) {
            // Won't succeed on retry — drop and report.
            rejected.push({ kind, uuid: item.uuid, status: res.status });
            continue;
        }
        // 5xx — server is alive but failing. Keep this item and stop.
        halted = true;
        remaining.push(item);
    }

    writeJSON(key, remaining);
    return { syncedCount, remainingCount: remaining.length, halted };
}

export async function flush(serverUrl: string): Promise<FlushOutcome> {
    const rejected: FlushOutcome["rejected"] = [];

    // Drain creates first — needed-state may FK-reference them.
    const created = await drainQueue<PendingCreate>(
        CREATES_KEY,
        listCreates(),
        "create",
        (c) =>
            apiFetch(`${serverUrl}/groceries/create`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ uuid: c.uuid, name: c.name }),
            }),
        rejected,
    );

    // If creates halted on a transient failure, leave needed for the next pass
    // so we don't issue PUTs whose target rows haven't been created server-side.
    if (created.halted) {
        return {
            synced: created.syncedCount,
            rejected,
            remaining: created.remainingCount + listNeeded().length,
        };
    }

    const settled = await drainQueue<PendingNeeded>(
        NEEDED_KEY,
        listNeeded(),
        "needed",
        (op) =>
            apiFetch(`${serverUrl}/groceries/${op.uuid}/needed`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ needed: op.needed }),
            }),
        rejected,
    );

    return {
        synced: created.syncedCount + settled.syncedCount,
        rejected,
        remaining: created.remainingCount + settled.remainingCount,
    };
}
