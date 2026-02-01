export default async function ConfigsIndexPage() {
  const { ConfigsIndexClient } = await import("./_components/ConfigsIndexClient");
  return <ConfigsIndexClient />;
}

