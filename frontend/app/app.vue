<script setup>
import { onAuthStateChanged, signOut } from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();
const user = ref(null);

onMounted(() => {
  onAuthStateChanged($firebaseAuth, (currentUser) => {
    user.value = currentUser;
  });
});

async function logout() {
  await signOut($firebaseAuth);
}
</script>

<template>
  <div
    class="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-50"
  >
    <header
      class="p-4 border-b bg-white dark:bg-gray-800 flex justify-between items-center"
    >
      <NuxtLink to="/" class="text-xl font-bold text-primary"
        >🎬 MovieTicket</NuxtLink
      >
      <div class="flex gap-4 items-center">
        <template v-if="user">
          <span class="text-sm text-gray-500">{{ user.email }}</span>
          <UButton to="/admin/bookings" variant="ghost" size="sm"
            >Admin Panel</UButton
          >
          <UButton color="red" variant="ghost" size="sm" @click="logout"
            >Logout</UButton
          >
        </template>
        <template v-else>
          <UButton to="/login" variant="ghost">Login</UButton>
        </template>
      </div>
    </header>
    <main>
      <NuxtPage />
    </main>
  </div>
</template>
