import type { ConfigFormat } from "@/lib/configApi";

export function prettify(format: ConfigFormat, raw: string): string {
  if (format !== "json") return raw;
  try {
    const obj = JSON.parse(raw);
    return JSON.stringify(obj, null, 2) + "\n";
  } catch {
    return raw;
  }
}

export function defaultConfigBody(format: ConfigFormat): string {
  return format === "json" ? "{\n  \n}\n" : "key: value\n";
}

