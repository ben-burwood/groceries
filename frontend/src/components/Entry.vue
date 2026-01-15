<template>
    <div class="join">
        <input type="text" class="input w-full join-item" placeholder="New grocery..." v-model="text" @keyup.enter="addItem" />
        <button class="btn btn-accent join-item" @click="addItem" :disabled="text.trim() === ''">Add</button>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const emit = defineEmits(["add"]);

const text = ref("");

function addItem() {
    if (text.value.trim() === "") return;
    emit("add", text.value);
    text.value = "";
}
</script>
