import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import ActivityFeed from "$lib/components/ActivityFeed.svelte";
import CommunitySection from "$lib/components/CommunitySection.svelte";

describe("home sections initial data", () => {
  it("renders recent activity without waiting for client hydration", () => {
    render(ActivityFeed, {
      recentOnly: true,
      initialActivities: [
        {
          type: "follow",
          target_type: "user",
          target_id: "target-user",
          user: { name: "Alice" },
          target: { name: "Bob", slug: "bob" },
          created_at: new Date().toISOString(),
        },
      ],
    });

    expect(screen.getByText("What's Happening")).toBeInTheDocument();
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
  });

  it("renders partners without waiting for client hydration", () => {
    render(CommunitySection, {
      initialPartners: [
        {
          uuid: "partner-uuid",
          name: "Anime Community",
          url: "https://example.com",
          type: "community",
        },
      ],
    });

    expect(screen.getByText("Partners & Communities")).toBeInTheDocument();
    expect(screen.getByTitle("Anime Community")).toBeInTheDocument();
  });
});
