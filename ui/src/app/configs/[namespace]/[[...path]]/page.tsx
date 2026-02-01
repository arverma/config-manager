export default async function ConfigPage() {
  const { ConfigPageClient } = await import("../../_components/ConfigPageClient");
  return <ConfigPageClient />;
}

