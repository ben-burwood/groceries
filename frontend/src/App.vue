<template>
    <div class="min-h-screen min-w-screen bg-base-200">
        <ThemeSwitcher class="absolute top-4 right-4" />

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

const errorMessage = ref("");
// clear error after 5 seconds
watch(errorMessage, (newError) => {
    if (newError) {
        setTimeout(() => {
            errorMessage.value = "";
        }, 5000);
    }
});

const allGroceries = ref<{ uuid: string; name: string }[]>([]);
async function fetchAllGroceries() {
    try {
        const res = await fetch(`${SERVER_URL}/groceries`);
        allGroceries.value = await res.json();
    } catch (error) {
        errorMessage.value = `Error: Fetching Groceries : ${error.message}`;
    }
}
onMounted(fetchAllGroceries);

const neededGroceryUuids = ref<string[]>([]);
async function fetchNeededGroceries() {
    try {
        const res = await fetch(`${SERVER_URL}/groceries/needed`);
        neededGroceryUuids.value = await res.json();
    } catch (error) {
        errorMessage.value = `Error: Fetching Needed Groceries : ${error.message}`;
    }
}
onMounted(fetchNeededGroceries);

const neededGroceries = computed(() => allGroceries.value.filter((grocery) => neededGroceryUuids.value.includes(grocery.uuid)));
const possibleGroceries = computed(() => allGroceries.value.filter((grocery) => !neededGroceryUuids.value.includes(grocery.uuid)));

const groceryModal = ref<HTMLDialogElement | null>(null);

async function addGrocery(name: string) {
    try {
        const res = await fetch(`${SERVER_URL}/groceries/create`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name }),
        });
        groceryModal.value?.close();
        await fetchAllGroceries();
    } catch (error) {
        errorMessage.value = `Error: Adding Grocery : ${error.message}`;
    }
}

async function toggleNeeded(uuid: string) {
    try {
        const res = await fetch(`${SERVER_URL}/groceries/${uuid}/needed`, { method: "PUT" });
        await fetchNeededGroceries();
    } catch (error) {
        errorMessage.value = `Error: Toggling Needed : ${error.message}`;
    }
}
</script>
