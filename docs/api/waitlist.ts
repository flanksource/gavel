import { kv } from "@vercel/kv";
import { Resend } from "resend";
import { z } from "zod";

// Vercel Node serverless handler signature (loose typing avoids pulling @vercel/node as a direct dep).
interface VercelReq {
  method?: string;
  headers: Record<string, string | string[] | undefined>;
  body: unknown;
}
interface VercelRes {
  status: (code: number) => VercelRes;
  json: (body: unknown) => void;
  setHeader: (name: string, value: string) => void;
  end: () => void;
}

const WaitlistBody = z.object({
  email: z.string().email().max(254),
  company: z.string().trim().max(200).optional(),
  use_case: z.string().trim().max(500).optional(),
});

const RATE_WINDOW_MS = 60_000;
const RATE_MAX = 10;

// Per-process LRU for rate limiting. Resets on cold start; see README for production note.
const hits = new Map<string, number[]>();

function clientIp(req: VercelReq): string {
  const fwd = req.headers["x-forwarded-for"];
  if (typeof fwd === "string") return fwd.split(",")[0]!.trim();
  if (Array.isArray(fwd) && fwd[0]) return fwd[0];
  return "unknown";
}

function isRateLimited(ip: string): boolean {
  const now = Date.now();
  const recent = (hits.get(ip) ?? []).filter((t) => now - t < RATE_WINDOW_MS);
  if (recent.length >= RATE_MAX) {
    hits.set(ip, recent);
    return true;
  }
  recent.push(now);
  hits.set(ip, recent);
  return false;
}

function resendClient(): Resend {
  const key = process.env.RESEND_API_KEY;
  if (!key) throw new Error("RESEND_API_KEY is not configured");
  return new Resend(key);
}

function fromAddress(): string {
  return process.env.RESEND_FROM ?? "Gavel <hello@gavel.flanksource.com>";
}

export default async function handler(req: VercelReq, res: VercelRes) {
  if (req.method !== "POST") {
    res.setHeader("allow", "POST");
    res.status(405).json({ error: "Method not allowed" });
    return;
  }

  const ip = clientIp(req);
  if (isRateLimited(ip)) {
    res.status(429).json({ error: "Too many requests, please try again in a minute." });
    return;
  }

  const parsed = WaitlistBody.safeParse(
    typeof req.body === "string" ? safeJson(req.body) : req.body,
  );
  if (!parsed.success) {
    res.status(400).json({ error: "Invalid request body", details: parsed.error.issues });
    return;
  }
  const { email, company, use_case } = parsed.data;

  const audienceId = process.env.RESEND_AUDIENCE_ID;
  if (!audienceId) {
    res.status(500).json({ error: "Waitlist not configured (missing RESEND_AUDIENCE_ID)" });
    return;
  }

  try {
    // Source of truth — Vercel KV.
    const key = `waitlist:${email.toLowerCase()}`;
    const now = new Date().toISOString();
    await kv.hset(key, { email, company: company ?? "", use_case: use_case ?? "", created_at: now, ip });
    await kv.sadd("waitlist:emails", email.toLowerCase());

    // Marketing audience — Resend.
    const resend = resendClient();
    await resend.contacts.create({ audienceId, email, firstName: company ?? undefined });

    // Confirmation email.
    await resend.emails.send({
      from: fromAddress(),
      to: email,
      subject: "You're on the Gavel waitlist",
      text:
        "Thanks for signing up for the Gavel waitlist.\n\n" +
        "We'll email you when the hosted tier opens for preview. In the meantime, the OSS CLI is live at https://github.com/flanksource/gavel.\n\n" +
        "— The Flanksource team",
    });

    res.status(200).json({ ok: true });
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown error";
    res.status(500).json({ error: `Signup failed: ${message}` });
  }
}

function safeJson(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}
