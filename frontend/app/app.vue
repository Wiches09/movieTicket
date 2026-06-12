<script setup>
import { onAuthStateChanged, signOut } from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();
const user = ref(null);
const userRole = useState("userRole", () => null);
const roleCookie = useCookie("userRole", { maxAge: 60 * 60 * 24 * 7 }); // 1 week

onMounted(() => {
  onAuthStateChanged($firebaseAuth, async (currentUser) => {
    user.value = currentUser;
    if (currentUser) {
      // Sync profile and get role
      try {
        const token = await currentUser.getIdToken();
        const res = await fetch("http://127.0.0.1:8080/api/profile/save", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            display_name:
              currentUser.displayName || currentUser.email.split("@")[0],
          }),
        });
        const profile = await res.json();
        userRole.value = profile.role || "user";
        roleCookie.value = userRole.value; // Save to cookie
      } catch (e) {
        console.error("Failed to sync profile:", e);
      }
    } else {
      userRole.value = null;
      roleCookie.value = null; // Clear cookie
    }
  });
});

async function logout() {
  await signOut($firebaseAuth);
  userRole.value = null;
  roleCookie.value = null; // Clear cookie
}
</script>

<template>
  <div
    class="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-50"
  >
    <header
      class="p-4 border-b bg-white dark:bg-gray-800 flex justify-between items-center"
    >
      <NuxtLink
        :to="userRole === 'admin' ? '/admin/bookings' : '/'"
        class="text-xl font-bold text-primary"
        >MovieTicket</NuxtLink
      >
      <div class="flex gap-4 items-center">
        <template v-if="user">
          <span class="text-sm text-gray-500">{{ user.email }}</span>
          <UButton
            v-if="userRole === 'admin'"
            to="/admin/bookings"
            variant="ghost"
            size="sm"
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
