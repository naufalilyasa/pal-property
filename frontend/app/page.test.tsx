import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("Home", () => {
  it("renders the seller app foundation shell", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", {
        level: 1,
        name: /a calm workspace for sellers to prepare listing operations/i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText(/seller app foundation/i)).toBeInTheDocument();
    expect(screen.getByText(/dashboard-ready shell/i)).toBeInTheDocument();
    expect(screen.getByText(/listing create, edit, and image tools are live/i)).toBeInTheDocument();
  });
});
