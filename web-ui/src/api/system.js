/**
 * Health probe for the backend. The Go server exposes GET /hello.
 */
export async function pingHello(client) {
  const res = await client.get('/hello', { responseType: 'text' })
  return typeof res.data === 'string' ? res.data : String(res.data)
}
