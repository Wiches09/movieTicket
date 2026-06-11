<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <UApp>
    <header class="bg-white border-b border-slate-200 sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <router-link to="/" class="text-2xl font-bold text-primary-600">MovieTicket</router-link>
        <nav>
          <UButton
            to="/"
            variant="ghost"
            color="neutral"
            class="text-zinc-900 hover:text-neutral-100"
            >Movies</UButton
          >
          <UButton
            to="/"
            variant="ghost"
            color="neutral"
            class="text-zinc-900 hover:text-neutral-100  ml-2"
            >Login</UButton
          >
        </nav>
      </div>
    </header>

    <main class="min-h-[calc(100vh-64px)] bg-slate-50">
      <RouterView />
    </main>
  </UApp>
</template>

<style>
body {
  margin: 0;
  background-color: #f8fafc;
  color: #0f172a;
}
</style>
