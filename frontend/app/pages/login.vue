<script setup>
definePageMeta({
  middleware: ["user"],
});

import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
} from "firebase/auth";
const { $firebaseAuth } = useNuxtApp();
const router = useRouter();

const isLogin = ref(true);
const email = ref("");
const password = ref("");
const loading = ref(false);
const user = ref(null);

onMounted(() => {
  onAuthStateChanged($firebaseAuth, (currentUser) => {
    user.value = currentUser;
  });
});

async function handleSubmit() {
  loading.value = true;
  try {
    let userCredential;
    if (isLogin.value) {
      userCredential = await signInWithEmailAndPassword(
        $firebaseAuth,
        email.value,
        password.value,
      );
    } else {
      userCredential = await createUserWithEmailAndPassword(
        $firebaseAuth,
        email.value,
        password.value,
      );
    }

    const user = userCredential.user;

    // sync firebase profile to MongoDB
    const token = await user.getIdToken();
    const res = await fetch("http://127.0.0.1:8080/api/profile/save", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        display_name: user.displayName || email.value.split("@")[0],
      }),
    });

    const profile = await res.json();
    if (profile.role === "admin") {
      router.push("/admin/bookings");
    } else {
      router.push("/");
    }
  } catch (error) {
    alert(error.message);
  } finally {
    loading.value = false;
  }
}

async function handleLogout() {
  await signOut($firebaseAuth);
}
</script>

<template>
  <div class="flex-grow flex items-center justify-center p-4 bg-slate-950">
    <UCard
      class="w-full max-w-md shadow-2xl border-slate-800 bg-slate-900"
      :ui="{
        header: { base: 'border-b border-slate-800' },
        body: { base: 'space-y-6' },
      }"
    >
      <template #header>
        <h1 class="text-3xl font-bold text-center text-white py-2">
          {{ user ? "Welcome Back!" : isLogin ? "Login" : "Register" }}
        </h1>
      </template>

      <div v-if="user" class="text-center space-y-6 py-4">
        <p class="text-slate-300">
          Logged in as: <strong class="text-white">{{ user.email }}</strong>
        </p>
        <div class="flex flex-col gap-3">
          <UButton to="/" block variant="outline" color="primary">
            Go to Movies
          </UButton>
          <UButton color="red" block variant="ghost" @click="handleLogout">
            Logout
          </UButton>
        </div>
      </div>

      <form v-else @submit.prevent="handleSubmit" class="space-y-6 py-2">
        <UFormField
          label="Email"
          :ui="{ label: { base: 'text-slate-200 font-semibold' } }"
        >
          <UInput
            v-model="email"
            type="email"
            placeholder="email@example.com"
            size="lg"
            variant="subtle"
            class="w-full"
            :ui="{
              base: 'bg-slate-800/50 border-slate-700 text-white placeholder-slate-500 focus:ring-primary-500',
            }"
            required
          />
        </UFormField>

        <UFormField
          label="Password"
          :ui="{ label: { base: 'text-slate-200 font-semibold' } }"
        >
          <UInput
            v-model="password"
            type="password"
            placeholder="********"
            size="lg"
            variant="subtle"
            class="w-full"
            :ui="{
              base: 'bg-slate-800/50 border-slate-700 text-white placeholder-slate-500 focus:ring-primary-500',
            }"
            required
          />
        </UFormField>

        <div class="pt-2">
          <UButton
            type="submit"
            color="primary"
            block
            size="lg"
            class="font-bold text-lg py-3"
            :loading="loading"
          >
            {{ isLogin ? "Sign In" : "Create Account" }}
          </UButton>
        </div>

        <div class="text-center">
          <button
            type="button"
            class="text-primary-400 hover:text-primary-300 text-sm font-medium transition-colors"
            @click="isLogin = !isLogin"
          >
            {{
              isLogin
                ? "Don't have an account? Register"
                : "Already have an account? Login"
            }}
          </button>
        </div>
      </form>
    </UCard>
  </div>
</template>
