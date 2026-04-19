import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WaitlistForm from "./WaitlistForm";

describe("WaitlistForm", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("rejects an invalid email without calling the API", async () => {
    const user = userEvent.setup();
    render(<WaitlistForm />);

    await user.type(screen.getByLabelText(/email/i), "not-an-email");
    await user.click(screen.getByRole("button", { name: /notify me/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/valid email/i);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("shows the success state after a 200 response", async () => {
    const mockFetch = fetch as unknown as ReturnType<typeof vi.fn>;
    mockFetch.mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200 }));

    const user = userEvent.setup();
    render(<WaitlistForm />);

    await user.type(screen.getByLabelText(/email/i), "dev@example.com");
    await user.click(screen.getByRole("button", { name: /notify me/i }));

    expect(await screen.findByText(/you're on the list/i)).toBeInTheDocument();
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/waitlist",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("surfaces the server error message when the API fails", async () => {
    const mockFetch = fetch as unknown as ReturnType<typeof vi.fn>;
    mockFetch.mockResolvedValue(
      new Response(JSON.stringify({ error: "Rate limited, try later" }), { status: 429 }),
    );

    const user = userEvent.setup();
    render(<WaitlistForm />);

    await user.type(screen.getByLabelText(/email/i), "dev@example.com");
    await user.click(screen.getByRole("button", { name: /notify me/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/rate limited/i);
  });
});
