"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { queryKeys } from "@/lib/query/keys";

export function DashboardRefreshButton() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return (
    <Button
      data-testid="dashboard-refresh-button"
      onClick={() => {
        queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });
        router.refresh();
      }}
      type="button"
      variant="secondary"
    >
      Refresh listings
    </Button>
  );
}
