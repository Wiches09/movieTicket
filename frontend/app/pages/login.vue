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
  <div class="min-h-screen flex items-center justify-center p-4">
    <UCard class="w-full max-w-md">
      <template #header>
        <h1 class="text-2xl font-bold text-center">
          {{ user ? "Welcome Back!" : isLogin ? "Login" : "Register" }}
        </h1>
      </template>

      <div v-if="user" class="text-center space-y-4">
        <p>
          Logged in as: <strong>{{ user.email }}</strong>
        </p>
        <div class="flex flex-col gap-2">
          <UButton to="/" block variant="outline">Go to Movies</UButton>
          <UButton color="red" block @click="handleLogout">Logout</UButton>
        </div>
      </div>

      <form v-else @submit.prevent="handleSubmit" class="space-y-4">
        <UFormField label="Email">
          <UInput
            v-model="email"
            type="email"
            placeholder="email@example.com"
            required
          />
        </UFormField>

        <UFormField label="Password">
          <UInput
            v-model="password"
            type="password"
            placeholder="********"
            required
          />
        </UFormField>

        <UButton type="submit" color="primary" block :loading="loading">
          {{ isLogin ? "Sign In" : "Create Account" }}
        </UButton>

        <div class="text-center text-sm">
          <button
            type="button"
            class="text-primary hover:underline"
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
