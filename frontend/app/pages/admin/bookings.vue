<script setup>
import { onAuthStateChanged } from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();

const user = ref(null);
const bookings = ref([]);
const usersMap = ref({});
const moviesMap = ref({});
const loading = ref(true);
const error = ref(null);

const filterType = ref("user"); // "user" or "movie"
const searchQuery = ref("");
const sortOrder = ref("desc"); // "desc" or "asc"

const filterOptions = ref([
  { value: "user", label: "Customer" },
  { value: "movie", label: "Movie" },
]);

function toggleSort() {
  sortOrder.value = sortOrder.value === "desc" ? "asc" : "desc";
}

const filteredBookings = computed(() => {
  let result = bookings.value;

  // 1. Filter
  if (searchQuery.value.trim() !== "") {
    result = result.filter((b) => {
      const q = searchQuery.value.toLowerCase();
      if (filterType.value === "user") {
        const userName = (
          usersMap.value[b.user_id] ||
          b.user_id ||
          ""
        ).toLowerCase();
        return userName.includes(q);
      } else if (filterType.value === "movie") {
        const movieName = (
          moviesMap.value[b.movie_id] || String(b.movie_id)
        ).toLowerCase();
        return movieName.includes(q);
      }
      return true;
    });
  }

  // 2. Sort by created_at
  result = [...result].sort((a, b) => {
    const dateA = new Date(a.created_at).getTime();
    const dateB = new Date(b.created_at).getTime();

    if (sortOrder.value === "asc") {
      return dateA - dateB;
    } else {
      return dateB - dateA;
    }
  });

  return result;
});

async function fetchData() {
  if (!user.value) return;

  loading.value = true;
  error.value = null;
  try {
    const token = await user.value.getIdToken();
    const headers = { Authorization: `Bearer ${token}` };

    // Fetch Bookings, Users, and Movies in parallel
    const [bookingsRes, usersRes, moviesRes] = await Promise.all([
      fetch("http://127.0.0.1:8080/api/admin/bookings", { headers }),
      fetch("http://127.0.0.1:8080/api/admin/users", { headers }),
      fetch("http://127.0.0.1:8080/api/movies"), // Movies are public
    ]);

    if (!bookingsRes.ok || !usersRes.ok || !moviesRes.ok) {
      const errorData = await (
        bookingsRes.ok ? (usersRes.ok ? moviesRes : usersRes) : bookingsRes
      ).json();
      throw new Error(errorData.error || "Failed to fetch data");
    }

    const bookingsData = await bookingsRes.json();
    const usersData = await usersRes.json();
    const moviesData = await moviesRes.json();

    // Map User UID to Display Name
    const uMap = {};
    usersData.forEach((u) => {
      uMap[u.uid] = u.display_name || u.email || "Unknown User";
    });

    // Map Movie ID to Title
    const mMap = {};
    moviesData.forEach((m) => {
      mMap[m.id] = m.title || `Movie #${m.id}`;
    });

    bookings.value = bookingsData;
    usersMap.value = uMap;
    moviesMap.value = mMap;
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
  { id: "movie", accessorKey: "movie_id", header: "Movie" },
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
      <div class="flex gap-4 mb-4 items-center flex-wrap">
        <USelect
          v-model="filterType"
          :items="filterOptions"
          option-attribute="label"
          value-attribute="value"
          class="w-48"
          :z-index="1000"
        />
        <UInput
          v-model="searchQuery"
          icon="i-heroicons-magnifying-glass"
          :placeholder="
            filterType === 'user'
              ? 'Search by customer name...'
              : 'Search by movie title...'
          "
          class="w-64"
        />
        <div class="flex-grow"></div>
        <UButton
          :icon="
            sortOrder === 'desc'
              ? 'i-heroicons-bars-arrow-down'
              : 'i-heroicons-bars-arrow-up'
          "
          color="gray"
          variant="solid"
          @click="toggleSort"
        >
          {{ sortOrder === "desc" ? "Newest First" : "Oldest First" }}
        </UButton>
      </div>

      <UTable :data="filteredBookings" :columns="columns" :loading="loading">
        <template #user-cell="{ row }">
          <div class="font-medium text-gray-900 dark:text-white">
            {{ usersMap[row.original.user_id] || row.original.user_id }}
          </div>
        </template>
        <template #movie-cell="{ row }">
          <div class="font-medium">
            {{
              moviesMap[row.original.movie_id] || `ID: ${row.original.movie_id}`
            }}
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
