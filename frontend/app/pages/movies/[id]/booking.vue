<script setup>
const route = useRoute();
const router = useRouter();
const movieId = route.params.id;
const { $firebaseAuth } = useNuxtApp();

const showtime = ref("19:00");

// Fetch movie details
const { data: movie } = await useFetch(
  `http://127.0.0.1:8080/api/movies/${movieId}`,
);

// Fetch occupied seats
const { data: occupiedSeats, refresh: refreshOccupied } = await useFetch(
  `http://127.0.0.1:8080/api/bookings/occupied`,
  {
    params: {
      movie_id: movieId,
      showtime: showtime.value,
    },
  },
);

const selectedSeats = ref([]);
const seats = ["A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"];

function toggleSeat(seat) {
  if (occupiedSeats.value?.includes(seat)) return;

  if (selectedSeats.value.includes(seat)) {
    selectedSeats.value = selectedSeats.value.filter((s) => s !== seat);
  } else {
    selectedSeats.value.push(seat);
  }
}

async function confirmBooking() {
  const user = $firebaseAuth.currentUser;
  if (!user) {
    alert("Please sign in to book tickets.");
    return;
  }

  if (selectedSeats.value.length === 0) {
    alert("Please select at least one seat.");
    return;
  }

  try {
    const token = await user.getIdToken();
    const response = await fetch("http://127.0.0.1:8080/api/bookings", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        movie_id: parseInt(movieId),
        seats: selectedSeats.value,
        showtime: showtime.value,
      }),
    });

    if (response.ok) {
      alert("Booking confirmed! Enjoy your movie.");
      router.push("/");
    } else {
      const errorData = await response.json();
      alert(`Booking failed: ${errorData.error || "Unknown error"}`);
      refreshOccupied(); // Refresh seats if there was a conflict
    }
  } catch (error) {
    console.error("Booking error:", error);
    alert("An error occurred while processing your booking.");
  }
}
</script>

<template>
  <div class="p-8 max-w-4xl mx-auto">
    <UButton
      :to="`/movies/${movieId}`"
      icon="i-heroicons-arrow-left"
      variant="ghost"
      class="mb-6"
    >
      Back to Details
    </UButton>

    <UCard>
      <template #header>
        <h2 v-if="movie" class="text-2xl font-bold text-emerald-600">
          Booking for Movie: {{ movie.title }}
        </h2>
        <h2 v-else class="text-2xl font-bold">Loading Booking...</h2>
      </template>

      <div class="p-4">
        <p class="mb-4 font-medium">Select your seats ({{ showtime }}):</p>
        <div class="grid grid-cols-3 gap-4 max-w-xs mx-auto mb-8">
          <UButton
            v-for="seat in seats"
            :key="seat"
            :color="
              occupiedSeats?.includes(seat)
                ? 'red'
                : selectedSeats.includes(seat)
                  ? 'emerald'
                  : 'gray'
            "
            :variant="
              occupiedSeats?.includes(seat) || selectedSeats.includes(seat)
                ? 'solid'
                : 'outline'
            "
            :disabled="occupiedSeats?.includes(seat)"
            @click="toggleSeat(seat)"
          >
            {{ seat }}
          </UButton>
        </div>

        <div
          v-if="selectedSeats.length > 0"
          class="text-center text-sm text-gray-500"
        >
          Selected: {{ selectedSeats.join(", ") }}
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton
            color="primary"
            :disabled="selectedSeats.length === 0"
            @click="confirmBooking"
          >
            Confirm Booking
          </UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
