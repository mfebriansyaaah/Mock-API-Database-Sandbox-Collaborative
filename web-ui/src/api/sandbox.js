/**
 * Sandbox API wrappers. The backend Go service exposes:
 *   GET    /sandbox/:projectId
 *   GET    /sandbox/:projectId/:table[?limit=N&offset=N]
 *   GET    /sandbox/:projectId/:table/:id
 *   POST   /sandbox/:projectId/:table
 *   POST   /sandbox/:projectId/:table/:id
 *   PUT    /sandbox/:projectId/:table/:id
 *   PATCH  /sandbox/:projectId/:table/:id
 *   DELETE /sandbox/:projectId/:table/:id
 */

/**
 * Fetch a page of documents from a table.
 * Returns { data, limit, offset, count, nextOffset }.
 * If the backend is on the old shape (bare array), normalise to that same
 * envelope so callers always see the same structure.
 */
export async function getTableDocuments(client, projectId, table, { limit, offset } = {}) {
  const params = {}
  if (limit != null) params.limit = limit
  if (offset != null) params.offset = offset
  const res = await client.get(`/sandbox/${projectId}/${table}`, { params })
  // New wrapped response
  if (res.data && typeof res.data === 'object' && Array.isArray(res.data.data)) {
    return {
      data: res.data.data,
      limit: res.data.limit,
      offset: res.data.offset,
      count: res.data.count,
      nextOffset: res.data.nextOffset
    }
  }
  // Fallback: legacy bare array shape
  if (Array.isArray(res.data)) {
    return {
      data: res.data,
      limit: res.data.length,
      offset: 0,
      count: res.data.length,
      nextOffset: res.data.length
    }
  }
  return { data: [], limit: 0, offset: 0, count: 0, nextOffset: 0 }
}

export async function getDocument(client, projectId, table, id) {
  const res = await client.get(`/sandbox/${projectId}/${table}/${id}`)
  return res.data
}

export async function createDocument(client, projectId, table, data, id) {
  if (id) {
    const res = await client.post(`/sandbox/${projectId}/${table}/${id}`, data)
    return res.data
  }
  const res = await client.post(`/sandbox/${projectId}/${table}`, data)
  return res.data
}

export async function updateDocument(client, projectId, table, id, data) {
  const res = await client.put(`/sandbox/${projectId}/${table}/${id}`, data)
  return res.data
}

export async function patchDocument(client, projectId, table, id, data) {
  const res = await client.patch(`/sandbox/${projectId}/${table}/${id}`, data)
  return res.data
}

export async function deleteDocument(client, projectId, table, id) {
  const res = await client.delete(`/sandbox/${projectId}/${table}/${id}`)
  return res.data
}

/**
 * Send a raw request through the configured client.
 * Used by the REST Tester.
 */
export async function sendRequest(client, { method, path, body, headers }) {
  const config = {
    method,
    url: path,
    data: body && ['POST', 'PUT', 'PATCH'].includes(method.toUpperCase())
      ? body
      : undefined,
    headers,
    transformRequest: [(d) => (typeof d === 'string' ? d : JSON.stringify(d))]
  }
  const res = await client.request(config)
  return res
}
