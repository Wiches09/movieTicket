// frontend/plugins/firebase.client.ts
import { initializeApp, getApps, getApp } from "firebase/app";
import { getAuth } from "firebase/auth";

export default defineNuxtPlugin(() => {
  const firebaseConfig = {
    apiKey: "AIzaSyB-1PAVani5xqQ0WUOcCg8GZgi2gu0vOSU",
    authDomain: "movieticket-df802.firebaseapp.com",
    projectId: "movieticket-df802",
    storageBucket: "movieticket-df802.firebasestorage.app",
    messagingSenderId: "1023879298187",
    appId: "1:1023879298187:web:92dd45bad1c9626f8e688e",
  };

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
