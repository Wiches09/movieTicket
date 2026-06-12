export default defineNuxtRouteMiddleware(async (to) => {
  const userRole = useState("userRole");
  const roleCookie = useCookie("userRole");

  // 1. Server-side or immediate check using Cookie
  const currentRole = userRole.value || roleCookie.value;
  if (currentRole === "admin") {
    return navigateTo("/admin/bookings");
  }

  if (process.client) {
    const nuxtApp = useNuxtApp();
    const $firebaseAuth = nuxtApp.$firebaseAuth;

    if ($firebaseAuth) {
      // 2. Wait for Firebase to initialize
      await new Promise((resolve) => {
        const unsubscribe = $firebaseAuth.onAuthStateChanged((user) => {
          unsubscribe();
          resolve(user);
        });
      });

      const user = $firebaseAuth.currentUser;

      // 3. If logged in but no role, fetch it
      if (user && !userRole.value) {
        try {
          const token = await user.getIdToken();
          const res = await fetch("http://127.0.0.1:8080/api/profile/save", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
              display_name: user.displayName || user.email.split("@")[0],
            }),
          });
          if (res.ok) {
            const profile = await res.json();
            userRole.value = profile.role || "user";
          }
        } catch (e) {
          console.error("User middleware error:", e);
        }
      }

      // 4. If admin, redirect to admin dashboard
      if (userRole.value === "admin") {
        return navigateTo("/admin/bookings");
      }
    }
  }
});
