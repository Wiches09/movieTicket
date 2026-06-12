// frontend/plugins/firebase.client.ts
import { initializeApp, getApps, getApp } from "firebase/app";
import { getAuth } from "firebase/auth";

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  const firebaseConfig = config.public.firebase;

  // Prevent duplicate initialization instances during hot-reloads
  const app = getApps().length === 0 ? initializeApp(firebaseConfig) : getApp();

  // Bind auth
  const auth = getAuth(app);

  console.log("Firebase Auth online");

  return {
    provide: {
      firebaseAuth: auth,
    },
  };
});
