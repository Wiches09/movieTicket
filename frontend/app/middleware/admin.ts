export default defineNuxtRouteMiddleware(async (to) => {
  const userRole = useState("userRole");
  const roleCookie = useCookie("userRole");

  // 1. Server-side or immediate client-side check using Cookie
  const currentRole = userRole.value || roleCookie.value;
  if (currentRole && currentRole !== "admin") {
    return navigateTo("/");
  }

  // 2. Browser-only check for deeper verification
  if (process.client) {
    const { $firebaseAuth } = useNuxtApp();

    if ($firebaseAuth) {
      // Wait for Firebase to initialize
      await new Promise((resolve) => {
        const unsubscribe = $firebaseAuth.onAuthStateChanged((user) => {
          unsubscribe();
          resolve(user);
        });
      });

      const user = $firebaseAuth.currentUser;

      // If not logged in, go to login
      if (!user) {
        return navigateTo("/login");
      }

      // Final Role sync
      if (!userRole.value) {
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
          const profile = await res.json();
          userRole.value = profile.role || "user";
        } catch (e) {
          return navigateTo("/");
        }
      }

      if (userRole.value !== "admin") {
        return navigateTo("/");
      }
    }
  } else if (!currentRole) {
    // On server, if no cookie exists at all, we can't confirm admin,
    // so we let it pass to the client for Firebase to decide.
  }
});
