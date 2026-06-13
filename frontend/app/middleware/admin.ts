export default defineNuxtRouteMiddleware(async (to) => {
  const userRole = useState("userRole");
  const roleCookie = useCookie("userRole");

  // 1. Server-side or immediate client-side check using Cookie
  const currentRole = userRole.value || roleCookie.value;

  if (!process.client) {
    if (currentRole === "admin") return;
    return navigateTo(currentRole ? "/" : "/login");
  }

  if (process.client) {
    const { $firebaseAuth } = useNuxtApp();

    if ($firebaseAuth) {
      if (!userRole.value) {
        await new Promise((resolve) => {
          const unsubscribe = $firebaseAuth.onAuthStateChanged((user) => {
            unsubscribe();
            resolve(user);
          });
        });
      }

      const user = $firebaseAuth.currentUser;

      if (!user) {
        userRole.value = null;
        roleCookie.value = null;
        return navigateTo("/login");
      }

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
          roleCookie.value = userRole.value;
        } catch (e) {
          return navigateTo("/");
        }
      }

      if (userRole.value !== "admin") {
        return navigateTo("/");
      }
    }
  }
});
