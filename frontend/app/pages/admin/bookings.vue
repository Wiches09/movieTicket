<script setup>
import { onAuthStateChanged } from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();

const user = ref(null);
const bookings = ref([]);
const usersMap = ref({});
const loading = ref(true);
const error = ref(null);

async function fetchData() {
  if (!user.value) return;

  loading.value = true;
  error.value = null;
  try {
    const token = await user.value.getIdToken();
    const headers = { Authorization: `Bearer ${token}` };

    // Fetch Bookings and Users in parallel
    const [bookingsRes, usersRes] = await Promise.all([
      fetch("http://127.0.0.1:8080/api/admin/bookings", { headers }),
      fetch("http://127.0.0.1:8080/api/admin/users", { headers }),
    ]);

    if (!bookingsRes.ok || !usersRes.ok)
      throw new Error("Failed to fetch data");

    const bookingsData = await bookingsRes.json();
    const usersData = await usersRes.json();

    // Map User UID to Display Name
    const map = {};
    usersData.forEach((u) => {
      map[u.uid] = u.display_name || u.email || "Unknown User";
    });

    bookings.value = bookingsData;
    usersMap.value = map;
  } catch (err) {
    error.value = err.message;
    console.error(err);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  onAuthStateChanged($firebaseAuth, (currentUser) => {
    user.value = currentUser;
    if (currentUser) {
      fetchData();
    } else {
      loading.value = false;
    }
  });
});

const columns = [
  { id: "id", accessorKey: "id", header: "ID" },
  { id: "user", accessorKey: "user_id", header: "Customer" },
  { id: "movie_id", accessorKey: "movie_id", header: "Movie ID" },
  { id: "seats", accessorKey: "seats", header: "Seats" },
  { id: "showtime", accessorKey: "showtime", header: "Time" },
  { id: "status", accessorKey: "status", header: "Status" },
  { id: "created_at", accessorKey: "created_at", header: "Created At" },
];

function formatDate(dateString) {
  if (!dateString) return "-";
  return new Date(dateString).toLocaleString();
}
</script>

<template>
  <div class="p-8">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">Booking Management (Admin)</h1>
      <UButton
        icon="i-heroicons-arrow-path"
        :loading="loading"
        @click="fetchData"
      >
        Refresh
      </UButton>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-red-100 text-red-700 rounded-lg">
      Error: {{ error }}
    </div>

    <div
      v-if="!user && !loading"
      class="text-center p-8 border-2 border-dashed rounded-lg"
    >
      <p class="text-gray-500">
        Please login as an administrator to view booking history.
      </p>
      <UButton to="/login" class="mt-4">Go to Login</UButton>
    </div>

    <UCard v-else>
      <UTable :data="bookings" :columns="columns" :loading="loading">
        <template #user-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ usersMap[row.original.user_id] || row.original.user_id }}
          </div>
        </template>
        <template #seats-cell="{ row }">
          <UBadge
            color="gray"
            variant="soft"
            class="mr-1"
            v-for="seat in row.original.seats"
            :key="seat"
          >
            {{ seat }}
          </UBadge>
        </template>
        <template #created_at-cell="{ row }">
          {{ formatDate(row.original.created_at) }}
        </template>
        <template #status-cell="{ row }">
          <UBadge
            :color="row.original.status === 'confirmed' ? 'green' : 'red'"
          >
            {{ row.original.status }}
          </UBadge>
        </template>
      </UTable>
    </UCard>
  </div>
</template>
