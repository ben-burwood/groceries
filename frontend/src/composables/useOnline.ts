import { ref, onMounted, onUnmounted } from "vue";

export function useOnline() {
    const online = ref(typeof navigator !== "undefined" ? navigator.onLine : true);

    const goOnline = () => {
        online.value = true;
    };
    const goOffline = () => {
        online.value = false;
    };

    onMounted(() => {
        window.addEventListener("online", goOnline);
        window.addEventListener("offline", goOffline);
    });

    onUnmounted(() => {
        window.removeEventListener("online", goOnline);
        window.removeEventListener("offline", goOffline);
    });

    return { online };
}
