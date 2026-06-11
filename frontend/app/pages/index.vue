<script setup>
import { ref, onMounted } from "vue";
import { signInWithEmailAndPassword, signOut } from "firebase/auth";

// Form and UI States
const email = ref("");
const password = ref("");
const isLoading = ref(false);
const errorMessage = ref("");

// Response Data States
const userToken = ref(null);
const backendResponse = ref(null);

let authInstance = null;

onMounted(() => {
  try {
    const nuxtApp = useNuxtApp();
    if (nuxtApp && nuxtApp.$firebaseAuth) {
      authInstance = nuxtApp.$firebaseAuth;
      console.log(
        "✅ index.vue successfully linked to operational Firebase engine!",
      );
    }
  } catch (err) {
    console.error("Failed to read Nuxt Firebase Auth plugin:", err);
  }
});

async function login() {
  if (!email.value || !password.value) return;

  if (!authInstance) {
    errorMessage.value =
      "Authentication instance is still connecting. Please wait.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  backendResponse.value = null;

  try {
    // 1. Authenticate with Firebase Client SDK
    const userCredential = await signInWithEmailAndPassword(
      authInstance,
      email.value,
      password.value,
    );

    // 2. Grab the raw secure JWT Token string
    const token = await userCredential.user.getIdToken();
    userToken.value = token;

    // 3. Automatically save/update this user profile inside MongoDB via Go Backend
    await saveUserProfileToMongo(token);

    // 4. Run your secure endpoint verification check
    await testBackendConnection(token);
  } catch (error) {
    console.error("Firebase Login Error:", error);
    errorMessage.value = error.message || "Login failed.";
  } finally {
    isLoading.value = false;
  }
}

async function saveUserProfileToMongo(token) {
  try {
    // Split the email name to create a default Display Name placeholder string
    const defaultName = email.value.split("@")[0];

    const response = await $fetch("http://localhost:8080/api/profile/save", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`, // Pass token so Go Middleware can verify it
        "Content-Type": "application/json",
      },
      body: {
        display_name: defaultName, // Sends the text payload downstream
      },
    });
    console.log("MongoDB Sync Response:", response);
  } catch (err) {
    console.error("MongoDB Profile Sync Failed:", err);
    // Don't block the UI entirely, just log a warning for debugging
  }
}

async function testBackendConnection(token) {
  try {
    const response = await $fetch("http://localhost:8080/api/secure-data", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    backendResponse.value = response;
  } catch (err) {
    console.error("Backend Connection Error:", err);
    errorMessage.value =
      "Firebase authenticated, but Go Backend rejected the token or is offline.";
  }
}

async function logout() {
  if (authInstance) {
    await signOut(authInstance);
  }
  userToken.value = null;
  backendResponse.value = null;
  email.value = "";
  password.value = "";
}
</script>

<template>
  <div
    class="h-screen flex flex-col items-center justify-center p-6 bg-gray-50 dark:bg-gray-900"
  >
    <div class="w-full max-w-sm space-y-6">
      <client-only>
        <UCard v-if="!userToken">
          <template #header>
            <h3 class="text-xl font-bold text-center">Movie Ticket Auth</h3>
          </template>

          <form @submit.prevent="login" class="space-y-4">
            <UFormField label="Email">
              <UInput
                v-model="email"
                type="email"
                placeholder="you@example.com"
                icon="i-lucide-mail"
                class="w-full"
              />
            </UFormField>

            <UFormField label="Password">
              <UInput
                v-model="password"
                type="password"
                placeholder="••••••••"
                icon="i-lucide-lock"
                class="w-full"
              />
            </UFormField>

            <UButton type="submit" block color="primary" :loading="isLoading">
              Sign In
            </UButton>
          </form>

          <p
            v-if="errorMessage"
            class="text-sm text-red-500 mt-3 text-center font-medium"
          >
            {{ errorMessage }}
          </p>
        </UCard>

        <UCard v-else class="border border-emerald-500">
          <template #header>
            <div class="flex items-center justify-between">
              <span class="text-emerald-600 font-bold flex items-center gap-1">
                <UIcon name="i-lucide-circle-check" /> Authenticated
              </span>
              <UButton size="xs" color="gray" variant="ghost" @click="logout"
                >Logout</UButton
              >
            </div>
          </template>

          <div class="space-y-3">
            <div>
              <p class="text-xs text-gray-400 uppercase">
                Firebase Token Status
              </p>
              <p
                class="text-sm font-mono truncate bg-gray-100 dark:bg-gray-800 p-1.5 rounded mt-1"
              >
                {{ userToken }}
              </p>
            </div>

            <div v-if="backendResponse">
              <p class="text-xs text-emerald-500 uppercase font-semibold">
                Response from Go Echo Backend
              </p>
              <div
                class="bg-emerald-50 dark:bg-emerald-950/30 p-3 rounded mt-1 text-sm border border-emerald-200 dark:border-emerald-900"
              >
                <p class="font-medium text-emerald-900 dark:text-emerald-200">
                  {{ backendResponse.message }}
                </p>
                <p class="text-xs text-emerald-700 dark:text-emerald-400 mt-1">
                  Verified UID: {{ backendResponse.uid }}
                </p>
              </div>
            </div>
          </div>
        </UCard>
      </client-only>
    </div>
  </div>
</template>
