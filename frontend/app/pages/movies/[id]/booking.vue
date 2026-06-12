<script setup>
import { onAuthStateChanged } from "firebase/auth";
const route = useRoute();
const router = useRouter();
const movieId = route.params.id;
const { $firebaseAuth } = useNuxtApp();

const user = ref(null);
const showtime = ref("19:00");
const seats = ["A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"];

// Timers for seat locked by current user
const seatTimers = ref({});

onMounted(() => {
  onAuthStateChanged($firebaseAuth, (currentUser) => {
    user.value = currentUser;
  });
  connectWebSocket();
});

// Fetch movie details
const { data: movie } = await useFetch(
  `http://127.0.0.1:8080/api/movies/${movieId}`,
);

// Fetch occupancy
const { data: occupancy, refresh: refreshOccupancy } = await useFetch(
  `http://127.0.0.1:8080/api/bookings/occupied`,
  {
    params: {
      movie_id: movieId,
      showtime: showtime.value,
    },
    key: `occupancy-${movieId}-${showtime.value}-${Date.now()}`, // Force unique key
  },
);

let ws;
function connectWebSocket() {
  ws = new WebSocket("ws://127.0.0.1:8080/api/ws/bookings");
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    // SAFETY FALLBACK: Always trigger a full refresh from the API
    // whenever ANY booking-related message is received.
    refreshOccupancy();

    if (data.type === "SEAT_UPDATE") {
      // Only update if it's for the current movie and showtime
      if (data.movie_id == movieId && data.showtime == showtime.value) {
        occupancy.value = {
          booked: data.booked,
          locked: data.locked,
        };
      }
    } else if (data.type === "RELOAD_SEATS") {
      // Refresh occupancy already handled by fallback
    }
  };
  ws.onclose = () => {
    setTimeout(connectWebSocket, 3000);
  };
}

onUnmounted(() => {
  if (ws) ws.close();
  // Clear all intervals
  Object.values(seatTimers.value).forEach((t) => clearInterval(t.interval));
});

function getSeatStatus(seat) {
  if (occupancy.value?.booked?.includes(seat)) return "booked";
  const lockerUid = occupancy.value?.locked?.[seat];
  if (lockerUid) {
    if (lockerUid === user.value?.uid) return "selected";
    return "locked";
  }
  return "available";
}

function getSeatClass(seat) {
  const status = getSeatStatus(seat);
  const base =
    "h-16 w-full rounded-md font-bold transition-all flex flex-col items-center justify-center text-sm relative ";

  switch (status) {
    case "booked":
      return base + "bg-red-600 text-white cursor-not-allowed";
    case "locked":
      return base + "bg-amber-500 text-white cursor-not-allowed";
    case "selected":
      return (
        base +
        "bg-emerald-600 text-white cursor-pointer ring-4 ring-emerald-300 ring-inset"
      );
    default:
      return (
        base +
        "bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 border border-gray-300 dark:border-gray-700 hover:bg-gray-200 cursor-pointer"
      );
  }
}

function startTimer(seat) {
  if (seatTimers.value[seat]) {
    clearInterval(seatTimers.value[seat].interval);
  }

  seatTimers.value[seat] = {
    timeLeft: 300, // 5 minutes
    interval: setInterval(() => {
      if (seatTimers.value[seat].timeLeft > 0) {
        seatTimers.value[seat].timeLeft--;
      } else {
        clearInterval(seatTimers.value[seat].interval);
        delete seatTimers.value[seat];
        refreshOccupancy();
      }
    }, 1000),
  };
}

async function toggleSeat(seat) {
  const status = getSeatStatus(seat);
  if (!user.value) return alert("Please sign in first.");

  const token = await user.value.getIdToken();

  if (status === "selected") {
    const res = await fetch("http://127.0.0.1:8080/api/bookings/unlock", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        movie_id: parseInt(movieId),
        showtime: showtime.value,
        seat,
      }),
    });
    if (res.ok) {
      if (seatTimers.value[seat]) {
        clearInterval(seatTimers.value[seat].interval);
        delete seatTimers.value[seat];
      }
      refreshOccupancy();
    }
  } else if (status === "available") {
    const res = await fetch("http://127.0.0.1:8080/api/bookings/lock", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        movie_id: parseInt(movieId),
        showtime: showtime.value,
        seat,
      }),
    });

    if (res.ok) {
      startTimer(seat);
      refreshOccupancy();
    } else {
      const err = await res.json();
      alert(err.error || "Failed to lock seat");
      refreshOccupancy();
    }
  }
}

async function confirmBooking() {
  if (!user.value) return;
  const selectedSeats = Object.keys(occupancy.value?.locked || {}).filter(
    (seat) => occupancy.value.locked[seat] === user.value.uid,
  );

  if (selectedSeats.length === 0)
    return alert("Please select at least one seat.");

  router.push({
    path: "/payment",
    query: {
      movieId: movieId,
      movie: movie.value?.title,
      seats: selectedSeats.join(","),
      showtime: showtime.value,
    },
  });
}
</script>

<template>
  <div class="p-8 max-w-4xl mx-auto">
    <UButton
      :to="`/movies/${movieId}`"
      icon="i-heroicons-arrow-left"
      variant="ghost"
      class="mb-6"
      >Back</UButton
    >

    <UCard>
      <template #header>
        <h2 v-if="movie" class="text-2xl font-bold text-emerald-600">
          Booking: {{ movie.title }}
        </h2>
      </template>

      <div class="p-4">
        <div class="flex flex-wrap gap-4 mb-8 justify-center text-xs">
          <div class="flex items-center gap-1">
            <div class="w-4 h-4 bg-gray-100 border rounded"></div>
            Available
          </div>
          <div class="flex items-center gap-1">
            <div class="w-4 h-4 bg-amber-500 rounded"></div>
            Locked
          </div>
          <div class="flex items-center gap-1">
            <div class="w-4 h-4 bg-red-600 rounded"></div>
            Booked
          </div>
          <div class="flex items-center gap-1">
            <div class="w-4 h-4 bg-emerald-600 rounded"></div>
            Yours
          </div>
        </div>

        <div class="grid grid-cols-3 gap-6 max-w-sm mx-auto mb-10">
          <button
            v-for="seat in seats"
            :key="seat"
            :class="getSeatClass(seat)"
            @click="toggleSeat(seat)"
          >
            <span>{{ seat }}</span>
            <span
              v-if="seatTimers[seat]"
              class="text-[10px] mt-1 bg-black/20 px-1 rounded"
            >
              {{ seatTimers[seat].timeLeft }}s
            </span>
          </button>
        </div>

        <p class="text-center text-sm text-gray-500 italic">
          Tip: Seats auto-unlock after 5 minutes. Complete payment quickly!
        </p>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton color="primary" @click="confirmBooking"
            >Confirm Booking</UButton
          >
        </div>
      </template>
    </UCard>
  </div>
</template>
