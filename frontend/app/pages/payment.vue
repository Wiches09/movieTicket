<script setup>
import { onAuthStateChanged } from "firebase/auth";
const route = useRoute();
const router = useRouter();
const { $firebaseAuth } = useNuxtApp();

const user = ref(null);
const loading = ref(false);
const movieTitle = ref(route.query.movie || "Movie");
const seatList = ref(route.query.seats?.split(",") || []);
const movieId = route.query.movieId;
const showtime = route.query.showtime;

onMounted(() => {
  onAuthStateChanged($firebaseAuth, (currentUser) => {
    user.value = currentUser;
  });
});

async function processPayment() {
  if (!user.value) return;
  loading.value = true;

  try {
    const token = await user.value.getIdToken();
    const response = await fetch("http://127.0.0.1:8080/api/bookings", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        movie_id: parseInt(movieId),
        seats: seatList.value,
        showtime: showtime,
      }),
    });

    if (response.ok) {
      alert("Payment Successful! Your tickets are confirmed.");
      router.push("/");
    } else {
      const errorData = await response.json();
      alert(`Payment failed: ${errorData.error || "Please try again"}`);
    }
  } catch (error) {
    console.error("Payment error:", error);
    alert("An error occurred during payment.");
  } finally {
    loading.value = false;
  }
}

function cancelPayment() {
    router.back();
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gray-50 dark:bg-gray-900">
    <UCard class="w-full max-w-md">
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold text-emerald-600">Secure Payment</h1>
          <p class="text-sm text-gray-500">Complete your booking in 5 minutes</p>
        </div>
      </template>

      <div class="space-y-6">
        <div class="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg">
          <h3 class="font-bold mb-2">{{ movieTitle }}</h3>
          <div class="flex justify-between text-sm">
            <span>Showtime:</span>
            <span class="font-medium">{{ showtime }}</span>
          </div>
          <div class="flex justify-between text-sm mt-1">
            <span>Seats:</span>
            <span class="font-medium">{{ seatList.join(", ") }}</span>
          </div>
          <div class="border-t border-gray-300 dark:border-gray-700 mt-3 pt-3 flex justify-between font-bold text-lg">
            <span>Total:</span>
            <span>${{ seatList.length * 12 }}.00</span>
          </div>
        </div>

        <div class="space-y-4">
            <UFormField label="Card Number">
                <UInput placeholder="1234 5678 9101 1121" icon="i-heroicons-credit-card" />
            </UFormField>
            <div class="grid grid-cols-2 gap-4">
                <UFormField label="Expiry">
                    <UInput placeholder="MM/YY" />
                </UFormField>
                <UFormField label="CVC">
                    <UInput placeholder="123" />
                </UFormField>
            </div>
        </div>
      </div>

      <template #footer>
        <div class="flex flex-col gap-2">
          <UButton color="primary" block size="lg" :loading="loading" @click="processPayment">
            Pay Now
          </UButton>
          <UButton variant="ghost" block @click="cancelPayment">
            Cancel
          </UButton>
        </div>
      </template>
    </UCard>
  </div>
</template>
