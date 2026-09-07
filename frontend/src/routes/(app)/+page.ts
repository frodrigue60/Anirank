import api from "$lib/api";

export const load = async () => {
  const [homeResult, activitiesResult, partnersResult] =
    await Promise.allSettled([
      api.get("/home"),
      api.get("/activities/recent"),
      api.get("/partners"),
    ]);

  if (homeResult.status === "rejected") {
    console.error("Failed to load home page data", homeResult.reason);
  }
  if (activitiesResult.status === "rejected") {
    console.error("Failed to load recent activities", activitiesResult.reason);
  }
  if (partnersResult.status === "rejected") {
    console.error("Failed to load partners", partnersResult.reason);
  }

  return {
    homeData:
      homeResult.status === "fulfilled" ? homeResult.value.data.data : null,
    recentActivities:
      activitiesResult.status === "fulfilled"
        ? activitiesResult.value.data.data || []
        : [],
    partners:
      partnersResult.status === "fulfilled"
        ? partnersResult.value.data || []
        : [],
  };
};
