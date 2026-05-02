<template>
    <div class="min-h-screen min-w-screen bg-base-200">
        <div class="absolute top-4 right-4 flex items-center gap-2">
            <span v-if="!online" class="badge badge-warning">Offline</span>
            <span v-else-if="pendingCount > 0" class="badge badge-info">Syncing {{ pendingCount }}…</span>
            <ThemeSwitcher />
        </div>

        <div class="flex flex-col items-center justify-center p-5 w-full max-w-lg mx-auto">
            <h1 class="text-4xl font-bold">Groceries</h1>

            <!-- Add new Grocery -->
            <div class="fab">
                <button class="btn btn-lg btn-circle btn-accent" onclick="add_grocery_modal.showModal()">
                    <svg
                        aria-label="New"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="2"
                        stroke="currentColor"
                        class="size-6"
                    >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                    </svg>
                </button>
            </div>
            <dialog ref="groceryModal" id="add_grocery_modal" class="modal">
                <div class="modal-box">
                    <h3 class="text-lg font-bold">Add a new Grocery</h3>
                    <Entry @add="addGrocery" class="mt-5 w-full" />
                </div>
                <form method="dialog" class="modal-backdrop">
                    <button>close</button>
                </form>
            </dialog>

            <!-- Grocery List -->
            <div class="flex flex-col w-full mt-5">
                <span v-if="neededGroceries.length === 0" class="text-4xl text-center">🎉</span>
                <div v-else>
                    <div class="grid grid-cols-3 gap-2 mt-2">
                        <Grocery
                            v-for="grocery in neededGroceries"
                            :key="grocery.uuid"
                            :groceryName="grocery.name"
                            :class="{ 'opacity-60': grocery.pending }"
                            @toggle="toggleNeeded(grocery.uuid)"
                        />
                    </div>
                </div>

                <div class="divider my-5"></div>

                <div class="grid grid-cols-3 gap-2 mt-2">
                    <Grocery
                        v-for="grocery in possibleGroceries"
                        :key="grocery.uuid"
                        :groceryName="grocery.name"
                        :class="{ 'opacity-60': grocery.pending }"
                        @toggle="toggleNeeded(grocery.uuid)"
                    />
                </div>
            </div>

            <!-- Error Toast -->
            <div v-if="errorMessage" role="alert" class="alert alert-error fixed bottom-10">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                </svg>
                {{ errorMessage }}
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import Grocery from "@/components/Grocery.vue";
import Entry from "@/components/Entry.vue";
import ThemeSwitcher from "@/components/ThemeSwitcher.vue";
import { SERVER_URL } from "@/main";
import * as offlineQueue from "@/services/offlineQueue";
import * as cache from "@/services/groceriesCache";
import { useOnline } from "@/composables/useOnline";
import type { Grocery as GroceryType, GroceryItem } from "@/types/grocery";

const errorMessage = ref("");
watch(errorMessage, (newError) => {
    if (newError) {
        setTimeout(() => {
            errorMessage.value = "";
        }, 5000);
    }
});

const { online } = useOnline();
const allGroceries = ref<GroceryItem[]>([]);
const neededUuids = ref<Set<string>>(new Set());
const pendingCount = ref(0);

const neededGroceries = computed(() => allGroceries.value.filter((g) => neededUuids.value.has(g.uuid)));
const possibleGroceries = computed(() => allGroceries.value.filter((g) => !neededUuids.value.has(g.uuid)));

function mergePending(serverList: GroceryType[], pendingCreates: { uuid: string; name: string }[]): GroceryItem[] {
    const seen = new Set(serverList.map((g) => g.uuid));
    const known = serverList.map<GroceryItem>((g) => ({ ...g, needed: neededUuids.value.has(g.uuid) }));
    const stillPending = pendingCreates
        .filter((c) => !seen.has(c.uuid))
        .map<GroceryItem>((c) => ({ ...c, needed: neededUuids.value.has(c.uuid), pending: true }));
    return [...known, ...stillPending];
}

async function fetchGroceries() {
    try {
        const [allRes, needRes] = await Promise.all([fetch(`${SERVER_URL}/groceries`), fetch(`${SERVER_URL}/groceries/needed`)]);
        if (!allRes.ok || !needRes.ok) throw new Error(`HTTP ${allRes.status}/${needRes.status}`);
        const fresh: GroceryType[] = await allRes.json();
        const needed: string[] = await needRes.json();
        cache.saveAll(fresh);
        cache.saveNeeded(needed);
        const queue = offlineQueue.snapshot();
        neededUuids.value = offlineQueue.applyNeededOverlay(needed, queue.needed);
        allGroceries.value = mergePending(fresh, queue.creates);
        pendingCount.value = offlineQueue.totalPending(queue);
    } catch (error: any) {
        if (!online.value) return;
        errorMessage.value = `Error: Fetching Groceries : ${error.message}`;
    }
}

async function flushQueue() {
    if (!online.value) return;
    const outcome = await offlineQueue.flush(SERVER_URL);
    pendingCount.value = outcome.remaining;
    if (outcome.rejected.length > 0) {
        errorMessage.value = `Sync rejected ${outcome.rejected.length} op(s)`;
    }
}

async function syncWithServer() {
    await flushQueue();
    await fetchGroceries();
}

const groceryModal = ref<HTMLDialogElement | null>(null);

async function addGrocery(name: string) {
    const item: GroceryItem = { uuid: crypto.randomUUID(), name, needed: true, pending: true };
    offlineQueue.enqueueCreate({ uuid: item.uuid, name: item.name });
    offlineQueue.enqueueNeeded({ uuid: item.uuid, needed: true });
    neededUuids.value = new Set([...neededUuids.value, item.uuid]);
    allGroceries.value = [...allGroceries.value, item];
    pendingCount.value = offlineQueue.totalPending();
    groceryModal.value?.close();
    if (online.value) await syncWithServer();
}

async function toggleNeeded(uuid: string) {
    const desired = !neededUuids.value.has(uuid);
    offlineQueue.enqueueNeeded({ uuid, needed: desired });
    const next = new Set(neededUuids.value);
    if (desired) next.add(uuid);
    else next.delete(uuid);
    neededUuids.value = next;
    pendingCount.value = offlineQueue.totalPending();
    if (online.value) await syncWithServer();
}

watch(online, (isOnline) => {
    if (isOnline) syncWithServer();
});

onMounted(async () => {
    // Hydrate from cache first so the UI paints instantly even when offline.
    const queue = offlineQueue.snapshot();
    neededUuids.value = offlineQueue.applyNeededOverlay(cache.loadNeeded(), queue.needed);
    allGroceries.value = mergePending(cache.loadAll(), queue.creates);
    pendingCount.value = offlineQueue.totalPending(queue);
    if (online.value) await syncWithServer();
});
</script>
