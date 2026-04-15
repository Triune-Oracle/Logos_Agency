import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Home from "../Home";

describe("Home", () => {
  it("renders hero headline", () => {
    render(<Home />);
    expect(screen.getByText(/Complexity Into/i)).toBeInTheDocument();
  });

  it("renders navigation links", () => {
    render(<Home />);
    expect(screen.getByText("Philosophy")).toBeInTheDocument();
    expect(screen.getByText("Services")).toBeInTheDocument();
    expect(screen.getByText("Strategy")).toBeInTheDocument();
    expect(screen.getByText("Clientele")).toBeInTheDocument();
  });

  it("renders all six rituals of traction", () => {
    render(<Home />);
    const rituals = ["Extraction", "Coherence", "Crucible", "Talisman", "Deployment", "Eternalization"];
    rituals.forEach((ritual) => {
      expect(screen.getByText(ritual)).toBeInTheDocument();
    });
  });
});
