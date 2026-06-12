<script setup>
const movie = ref({
  title: '',
  year: new Date().getFullYear()
})

const isLoading = ref(false)
const message = ref('')
const isError = ref(false)

async function addMovie() {
  if (!movie.value.title) return

  isLoading.value = true
  message.value = ''

  try {
    const response = await $fetch('http://127.0.0.1:8080/api/movies', {
      method: 'POST',
      body: movie.value
    })

    message.value = `Successfully added: ${response.title}`
    isError.value = false

    // Clear form
    movie.value.title = ''
    movie.value.year = new Date().getFullYear()
  } catch (err) {
    console.error(err)
    message.value = 'Failed to add movie. Ensure backend is running.'
    isError.value = true
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="p-8 max-w-md mx-auto">
    <UCard>
      <template #header>
        <h1 class="text-xl font-bold">Add New Movie</h1>
      </template>

      <form @submit.prevent="addMovie" class="space-y-4">
        <UFormField label="Movie Title">
          <UInput v-model="movie.title" placeholder="e.g. Interstellar" required />
        </UFormField>

        <UFormField label="Release Year">
          <UInput v-model="movie.year" type="number" required />
        </UFormField>

        <UButton type="submit" block color="primary" :loading="isLoading">
          Add Movie
        </UButton>
      </form>

      <div v-if="message" :class="['mt-4 p-3 rounded text-sm text-center', isError ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600']">
        {{ message }}
      </div>

      <template #footer>
        <UButton to="/" variant="ghost" color="gray" block>
          Back to Home
        </UButton>
      </template>
    </UCard>
  </div>
</template>
