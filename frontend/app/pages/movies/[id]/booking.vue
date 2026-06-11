<script setup>
const route = useRoute();
const movieId = route.params.id;

// Fetch movie details from the backend to get the title
const { data: movie } = await useFetch(
  `http://127.0.0.1:8080/api/movies/${movieId}`,
);
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

      <div
        class="h-64 flex flex-col items-center justify-center border-2 border-dashed border-gray-200 dark:border-gray-800 rounded-lg"
      >
        <UIcon name="i-heroicons-ticket" class="text-4xl text-gray-300 mb-2" />
        <p class="text-gray-400 italic">
          Seat selection for "{{ movie?.title }}" will be implemented here.
        </p>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton color="primary" disabled> Confirm Booking </UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
