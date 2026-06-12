// frontend/plugins/firebase.client.ts
import { initializeApp, getApps, getApp } from "firebase/app";
import { getAuth } from "firebase/auth";

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  const firebaseConfig = {
    apiKey: config.public.firebaseApiKey,
    authDomain: config.public.firebaseAuthDomain,
    projectId: config.public.firebaseProjectId,
    storageBucket: config.public.firebaseStorageBucket,
    messagingSenderId: config.public.firebaseMessagingSenderId,
    appId: config.public.firebaseAppId,
  };

  // Prevent duplicate initialization instances during hot-reloads
  const app = getApps().length === 0 ? initializeApp(firebaseConfig) : getApp();

  // Bind auth
  const auth = getAuth(app);

  return {
    provide: {
      firebaseAuth: auth,
    },
  };
});
