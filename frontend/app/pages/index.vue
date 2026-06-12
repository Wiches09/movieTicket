<script setup>
definePageMeta({
  middleware: ["user"],
});

const {
  data: movies,
  pending,
  error,
  refresh,
} = await useFetch("http://127.0.0.1:8080/api/movies");
</script>

<template>
  <div class="p-8 max-w-6xl mx-auto">
    <header class="mb-10 text-center">
      <h1 class="text-4xl font-extrabold text-gray-900 dark:text-white">
        Now Showing
      </h1>
      <p class="text-gray-500 mt-2">
        Book your tickets for the latest blockbusters
      </p>
    </header>

    <div v-if="pending" class="flex justify-center py-20">
      <p>Loading movies...</p>
    </div>

    <div v-else-if="error" class="text-center py-20">
      <p class="text-red-500">Failed to load movies. Is the backend running?</p>
      <UButton @click="refresh" class="mt-4">Retry</UButton>
    </div>

    <div
      v-else
      class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6"
    >
      <UCard
        v-for="movie in movies"
        :key="movie.id"
        class="overflow-hidden hover:shadow-xl transition-shadow cursor-pointer"
        @click="navigateTo(`/movies/${movie.id}`)"
      >
        <template #header>
          <div
            class="h-48 bg-gray-200 dark:bg-gray-800 flex items-center justify-center"
          >
            <UIcon name="i-heroicons-film" class="text-4xl text-gray-400" />
          </div>
        </template>

        <div class="p-2 text-center">
          <h3 class="font-bold text-lg truncate">{{ movie.title }}</h3>
          <p class="text-sm text-gray-500">{{ movie.year }}</p>
        </div>

        <template #footer>
          <UButton
            block
            color="primary"
            variant="soft"
            :to="`/movies/${movie.id}`"
          >
            View Details
          </UButton>
        </template>
      </UCard>
    </div>
  </div>
</template>
