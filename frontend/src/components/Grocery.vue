<template>
    <div class="card bg-primary/40 hover:bg-secondary/40 shadow-lg" @click="emit('toggle')">
        <div class="card-body">
            <span class="card-title">{{ groceryName }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
defineProps<{
    groceryName: String;
}>();

const emit = defineEmits(["toggle"]);
</script>
