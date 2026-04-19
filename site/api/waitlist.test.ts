import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@vercel/kv", () => ({
  kv: {
    hset: vi.fn().mockResolvedValue(1),
    sadd: vi.fn().mockResolvedValue(1),
  },
}));

vi.mock("resend", () => ({
  Resend: vi.fn().mockImplementation(() => ({
    contacts: { create: vi.fn().mockResolvedValue({ data: { id: "c_1" } }) },
    emails: { send: vi.fn().mockResolvedValue({ data: { id: "e_1" } }) },
  })),
}));

import handler from "./waitlist";
import { kv } from "@vercel/kv";
import { Resend } from "resend";

interface MockRes {
  statusCode: number;
  body: unknown;
  headers: Record<string, string>;
  status(code: number): MockRes;
  json(body: unknown): void;
  setHeader(name: string, value: string): void;
  end(): void;
}

function makeRes(): MockRes {
  const res: MockRes = {
    statusCode: 0,
    body: undefined,
    headers: {},
    status(code) {
      this.statusCode = code;
      return this;
    },
    json(body) {
      this.body = body;
    },
    setHeader(name, value) {
      this.headers[name.toLowerCase()] = value;
    },
    end() {},
  };
  return res;
}

function makeReq(overrides: Partial<{ method: string; body: unknown; ip: string }> = {}) {
  return {
    method: overrides.method ?? "POST",
    headers: { "x-forwarded-for": overrides.ip ?? `10.0.0.${Math.floor(Math.random() * 254)}` },
    body: overrides.body ?? { email: "dev@example.com" },
  };
}

describe("api/waitlist handler", () => {
  beforeEach(() => {
    process.env.RESEND_API_KEY = "test-key";
    process.env.RESEND_AUDIENCE_ID = "aud_123";
    vi.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.RESEND_API_KEY;
    delete process.env.RESEND_AUDIENCE_ID;
  });

  it("rejects non-POST requests with 405", async () => {
    const res = makeRes();
    await handler(makeReq({ method: "GET" }), res);
    expect(res.statusCode).toBe(405);
    expect(res.headers.allow).toBe("POST");
  });

  it("returns 400 for an invalid email", async () => {
    const res = makeRes();
    await handler(makeReq({ body: { email: "not-an-email" } }), res);
    expect(res.statusCode).toBe(400);
  });

  it("stores the signup in KV and fires Resend on valid input", async () => {
    const res = makeRes();
    await handler(
      makeReq({ ip: "10.1.2.3", body: { email: "dev@example.com", company: "Acme" } }),
      res,
    );

    expect(res.statusCode).toBe(200);
    expect(kv.hset).toHaveBeenCalledWith(
      "waitlist:dev@example.com",
      expect.objectContaining({ email: "dev@example.com", company: "Acme" }),
    );
    expect(kv.sadd).toHaveBeenCalledWith("waitlist:emails", "dev@example.com");

    const ResendMock = Resend as unknown as ReturnType<typeof vi.fn>;
    const instance = ResendMock.mock.results[0]!.value as {
      contacts: { create: ReturnType<typeof vi.fn> };
      emails: { send: ReturnType<typeof vi.fn> };
    };
    expect(instance.contacts.create).toHaveBeenCalledWith(
      expect.objectContaining({ audienceId: "aud_123", email: "dev@example.com" }),
    );
    expect(instance.emails.send).toHaveBeenCalled();
  });

  it("rate-limits a single IP after 10 requests in the window", async () => {
    const ip = "10.9.9.9";
    let last: MockRes | null = null;
    for (let i = 0; i < 11; i += 1) {
      last = makeRes();
      await handler(makeReq({ ip, body: { email: `dev${i}@example.com` } }), last);
    }
    expect(last?.statusCode).toBe(429);
  });
});
