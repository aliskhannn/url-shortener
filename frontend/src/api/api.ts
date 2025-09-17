import type { Link } from "../entities/link";

export interface AnalyticsSummary {
  alias: string;
  total_clicks: number;
  daily: Record<string, number>;
  user_agent: Record<string, number>;
}

export async function createLink(req: {
  url: string;
  alias?: string;
}): Promise<Link> {
  const res = await fetch("http://localhost:8080/api/shorten", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error("Failed to create link");
  const data = await res.json();
  return data.result;
}

export async function getAnalytics(alias: string): Promise<AnalyticsSummary> {
  const res = await fetch(`http://localhost:8080/api/analytics/${alias}`);
  if (!res.ok) throw new Error("Failed to fetch analytics");
  const data = await res.json();
  return data; // возвращаем объект агрегированной статистики
}
