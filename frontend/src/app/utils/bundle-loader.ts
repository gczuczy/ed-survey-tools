/**
 * Fetch a gzip-compressed JSON bundle and return its parsed contents.
 * The server sets Content-Encoding: gzip, so the browser decompresses
 * the response automatically before presenting the body to JavaScript.
 */
export async function fetchBundle<T>(url: string): Promise<T[]> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(
      `Bundle fetch failed: ${resp.status} ${resp.statusText}`,
    );
  }
  return resp.json() as Promise<T[]>;
}
