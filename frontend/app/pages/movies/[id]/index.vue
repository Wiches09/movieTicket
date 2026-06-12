<script setup>
definePageMeta({
  middleware: ["user"],
});

const route = useRoute();
const movieId = route.params.id;

// Fetch movie details from the backend
const {
  data: movie,
  pending,
  error,
} = await useFetch(`http://127.0.0.1:8080/api/movies/${movieId}`);
</script>

<template>
  <div class="p-8 max-w-4xl mx-auto">
    <UButton to="/" icon="i-heroicons-arrow-left" variant="ghost" class="mb-6">
      Back to Movies
    </UButton>

    <div v-if="pending" class="flex justify-center py-20">
      <p>Loading movie details...</p>
    </div>

    <div v-else-if="error" class="text-center py-20">
      <UAlert
        color="red"
        variant="soft"
        title="Error"
        description="Could not find the movie or the backend is offline."
      />
    </div>

    <UCard v-else-if="movie" class="overflow-hidden">
      <div class="md:flex">
        <div
          class="md:w-1/3 bg-gray-200 dark:bg-gray-800 h-64 md:h-auto flex items-center justify-center"
        >
          <UIcon name="i-heroicons-film" class="text-6xl text-gray-400" />
        </div>

        <div class="md:w-2/3 p-6 space-y-4">
          <div>
            <h1 class="text-3xl font-bold">{{ movie.title }}</h1>
            <p class="text-gray-500">Release Year: {{ movie.year }}</p>
          </div>

          <div class="pt-4 border-t border-gray-100 dark:border-gray-800">
            <p class="text-gray-600 dark:text-gray-300">
              This is a placeholder description for "{{ movie.title }}". You can
              update your MongoDB document to include more details!
            </p>
          </div>

          <div class="pt-6">
            <UButton
              :to="`/movies/${movie.id}/booking`"
              size="xl"
              block
              color="primary"
            >
              Book Tickets
            </UButton>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>
