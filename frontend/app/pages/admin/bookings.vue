<script setup>
definePageMeta({
  middleware: ["admin"],
});
const userRole = useState("userRole");
import { onAuthStateChanged } from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();

const user = ref(null);
const timeline = ref([]);
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

const filteredTimeline = computed(() => {
  let result = timeline.value;

  // 1. Filter
  if (searchQuery.value.trim() !== "") {
    result = result.filter((item) => {
      const q = searchQuery.value.toLowerCase();
      if (filterType.value === "user") {
        const userName = (
          usersMap.value[item.user_id] ||
          item.user_id ||
          ""
        ).toLowerCase();
        return (
          userName.includes(q) ||
          (item.message && item.message.toLowerCase().includes(q))
        );
      } else if (filterType.value === "movie") {
        const movieName = (
          moviesMap.value[item.movie_id] || String(item.movie_id)
        ).toLowerCase();
        return (
          movieName.includes(q) ||
          (item.message && item.message.toLowerCase().includes(q))
        );
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

    // Fetch Bookings, Users, Movies, and Logs in parallel
    const [bookingsRes, usersRes, moviesRes, logsRes] = await Promise.all([
      fetch("http://127.0.0.1:8080/api/admin/bookings", { headers }),
      fetch("http://127.0.0.1:8080/api/admin/users", { headers }),
      fetch("http://127.0.0.1:8080/api/movies"), // Movies are public
      fetch("http://127.0.0.1:8080/api/admin/logs", { headers }),
    ]);

    if (!bookingsRes.ok || !usersRes.ok || !moviesRes.ok || !logsRes.ok) {
      throw new Error("Failed to fetch dashboard data");
    }

    const bookingsData = await bookingsRes.json();
    const usersData = await usersRes.json();
    const moviesData = await moviesRes.json();
    const logsData = await logsRes.json();

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

    usersMap.value = uMap;
    moviesMap.value = mMap;

    // Merge Bookings and Logs into a single Timeline array
    const merged = [];

    if (Array.isArray(bookingsData)) {
      bookingsData.forEach((b) => {
        merged.push({
          id: b.id,
          user_id: b.user_id,
          movie_id: b.movie_id,
          seats: b.seats,
          showtime: b.showtime,
          status: b.status,
          message: `Booked seats ${b.seats.join(", ")}`,
          created_at: b.created_at,
        });
      });
    }

    if (Array.isArray(logsData)) {
      logsData.forEach((l) => {
        // Skip successful booking logs to avoid duplication with actual bookings
        if (l.event_type !== "BOOKING SUCCESS") {
          merged.push({
            id: l.id,
            user_id: null,
            movie_id: null,
            seats: [],
            showtime: "-",
            status: l.event_type,
            message: l.message,
            created_at: l.created_at,
          });
        }
      });
    }

    timeline.value = merged;
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
  { id: "customer", accessorKey: "user_id", header: "Customer / Context" },
  { id: "movie", accessorKey: "movie_id", header: "Movie" },
  { id: "activity", accessorKey: "message", header: "Activity" },
  { id: "status", accessorKey: "status", header: "Status" },
  { id: "created_at", accessorKey: "created_at", header: "Time" },
];

function formatDate(dateString) {
  if (!dateString) return "-";
  return new Date(dateString).toLocaleString();
}
</script>

<template>
  <ClientOnly>
    <div v-if="userRole === 'admin'" class="p-8">
      <div v-if="error" class="mb-4 p-4 bg-red-100 text-red-700 rounded-lg">
        Error: {{ error }}
      </div>

      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h1 class="text-2xl font-bold">Booking Management (Admin)</h1>
            <UButton
              icon="i-heroicons-arrow-path"
              :loading="loading"
              @click="fetchData"
            >
              Refresh
            </UButton>
          </div>
        </template>

        <div class="flex gap-4 mb-4 items-center flex-wrap">
          <USelect
            v-model="filterType"
            :items="filterOptions"
            option-attribute="label"
            value-attribute="value"
            class="w-48"
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
            color="neutral"
            variant="outline"
            @click="toggleSort"
          >
            {{ sortOrder === "desc" ? "Newest First" : "Oldest First" }}
          </UButton>
        </div>

        <UTable :data="filteredTimeline" :columns="columns" :loading="loading">
          <template #customer-cell="{ row }">
            <div class="font-medium text-gray-900 dark:text-white">
              <span v-if="row.original.user_id">
                {{ usersMap[row.original.user_id] || row.original.user_id }}
              </span>
              <span v-else class="text-gray-500 italic">System Event</span>
            </div>
          </template>
          <template #movie-cell="{ row }">
            <div class="font-medium">
              <span v-if="row.original.movie_id">
                {{
                  moviesMap[row.original.movie_id] ||
                  `ID: ${row.original.movie_id}`
                }}
              </span>
              <span v-else>-</span>
            </div>
          </template>
          <template #activity-cell="{ row }">
            {{ row.original.message }}
          </template>
          <template #created_at-cell="{ row }">
            {{ formatDate(row.original.created_at) }}
          </template>
          <template #status-cell="{ row }">
            <UBadge
              :color="
                row.original.status === 'confirmed'
                  ? 'green'
                  : row.original.status.includes('TIMEOUT')
                    ? 'amber'
                    : row.original.status.includes('ERROR')
                      ? 'red'
                      : 'gray'
              "
            >
              {{ row.original.status }}
            </UBadge>
          </template>
        </UTable>
      </UCard>
    </div>
    <div v-else class="flex items-center justify-center min-h-[60vh]">
      <UProgress />
    </div>
  </ClientOnly>
</template>
