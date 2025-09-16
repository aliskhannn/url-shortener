import type { Link } from "../entities/link";
import type { Analytics } from "../entities/analytics";

export async function createLink(req: {
  url: string;
  alias?: string;
}): Promise<Link> {
  const res = await fetch("/shorten", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error("Failed to create link");
  return res.json() as Promise<Link>;
}

export async function getAnalytics(alias: string): Promise<Analytics[]> {
  const res = await fetch(`/analytics/${alias}`);
  if (!res.ok) throw new Error("Failed to fetch analytics");
  return res.json() as Promise<Analytics[]>;
}
