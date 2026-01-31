/**
 * Multi-language HTTP code generators.
 *
 * Converts a captured proxy request into executable code snippets
 * for cURL, JavaScript (Fetch), Python (requests), and Go (net/http).
 */

export type CodeLanguage = 'curl' | 'fetch' | 'python' | 'go';

export interface CodeGenRequest {
  method: string;
  url: string;
  headers?: Record<string, string[]>;
  body?: string;
}

/** Unified entry — pick a language and get the code string. */
export function generateCode(lang: CodeLanguage, req: CodeGenRequest): string {
  switch (lang) {
    case 'curl':
      return generateCurl(req);
    case 'fetch':
      return generateFetch(req);
    case 'python':
      return generatePython(req);
    case 'go':
      return generateGo(req);
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Headers to skip — pseudo-headers, host (in URL), and encoding (transport). */
const SKIP_HEADERS = new Set([
  'host',
  'content-length',
  'transfer-encoding',
  'connection',
]);

function shouldSkipHeader(key: string): boolean {
  return key.startsWith(':') || SKIP_HEADERS.has(key.toLowerCase());
}

/** Flatten Record<string, string[]> to [key, value][] pairs, skipping noise. */
function flatHeaders(headers?: Record<string, string[]>): [string, string][] {
  if (!headers) return [];
  const result: [string, string][] = [];
  for (const [key, values] of Object.entries(headers)) {
    if (shouldSkipHeader(key)) continue;
    for (const v of values) {
      result.push([key, v]);
    }
  }
  return result;
}

/** True when the body looks like valid JSON. */
function isJsonBody(body: string | undefined): boolean {
  if (!body) return false;
  try {
    JSON.parse(body);
    return true;
  } catch {
    return false;
  }
}

function hasBody(method: string): boolean {
  return ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method);
}

/** Content-Type from headers (first match). */
function getContentType(headers?: Record<string, string[]>): string | undefined {
  if (!headers) return undefined;
  for (const [k, v] of Object.entries(headers)) {
    if (k.toLowerCase() === 'content-type' && v.length > 0) return v[0];
  }
  return undefined;
}

// ---------------------------------------------------------------------------
// cURL
// ---------------------------------------------------------------------------

export function generateCurl(req: CodeGenRequest): string {
  const parts: string[] = ['curl'];

  if (req.method !== 'GET') {
    parts.push(`-X ${req.method}`);
  }

  for (const [key, value] of flatHeaders(req.headers)) {
    parts.push(`-H '${key}: ${value.replace(/'/g, "'\\''")}'`);
  }

  if (req.body && hasBody(req.method)) {
    const escaped = req.body.replace(/'/g, "'\\''");
    parts.push(`-d '${escaped}'`);
  }

  parts.push(`'${req.url}'`);

  return parts.join(' \\\n  ');
}

// ---------------------------------------------------------------------------
// JavaScript Fetch
// ---------------------------------------------------------------------------

export function generateFetch(req: CodeGenRequest): string {
  const headers = flatHeaders(req.headers);
  const wantBody = req.body && hasBody(req.method);
  const ct = getContentType(req.headers);
  const jsonBody = isJsonBody(req.body);

  // Simple GET with no special headers
  if (req.method === 'GET' && headers.length === 0) {
    return `const response = await fetch('${req.url}');
const data = await response.json();`;
  }

  const lines: string[] = [];
  lines.push(`const response = await fetch('${req.url}', {`);
  lines.push(`  method: '${req.method}',`);

  if (headers.length > 0) {
    // Filter out Content-Type if we'll use JSON.stringify (fetch auto-sets it)
    const filteredHeaders = jsonBody && wantBody
      ? headers.filter(([k]) => k.toLowerCase() !== 'content-type')
      : headers;

    if (filteredHeaders.length > 0) {
      lines.push('  headers: {');
      for (const [key, value] of filteredHeaders) {
        lines.push(`    '${key}': '${value.replace(/'/g, "\\'")}',`);
      }
      lines.push('  },');
    }
  }

  if (wantBody) {
    if (jsonBody) {
      // Pretty-print the JSON for readability
      try {
        const parsed = JSON.parse(req.body!);
        const pretty = JSON.stringify(parsed, null, 2);
        lines.push(`  body: JSON.stringify(${pretty.replace(/\n/g, '\n  ')}),`);
      } catch {
        lines.push(`  body: ${JSON.stringify(req.body)},`);
      }
    } else if (ct?.includes('x-www-form-urlencoded')) {
      lines.push(`  body: '${req.body!.replace(/'/g, "\\'")}',`);
    } else {
      lines.push(`  body: ${JSON.stringify(req.body)},`);
    }
  }

  lines.push('});');
  lines.push('const data = await response.json();');

  return lines.join('\n');
}

// ---------------------------------------------------------------------------
// Python requests
// ---------------------------------------------------------------------------

export function generatePython(req: CodeGenRequest): string {
  const headers = flatHeaders(req.headers);
  const wantBody = req.body && hasBody(req.method);
  const jsonBody = isJsonBody(req.body);
  const ct = getContentType(req.headers);
  const methodLower = req.method.toLowerCase();

  const lines: string[] = ['import requests', ''];

  // Build the call
  const args: string[] = [];
  args.push(`    '${req.url}'`);

  // Headers — skip Content-Type when using json= kwarg
  const filteredHeaders = jsonBody && wantBody
    ? headers.filter(([k]) => k.toLowerCase() !== 'content-type')
    : headers;

  if (filteredHeaders.length > 0) {
    const hLines: string[] = ['{'];
    for (const [key, value] of filteredHeaders) {
      hLines.push(`        '${key}': '${value.replace(/'/g, "\\'")}',`);
    }
    hLines.push('    }');
    args.push(`    headers=${hLines.join('\n')}`);
  }

  if (wantBody) {
    if (jsonBody) {
      try {
        const parsed = JSON.parse(req.body!);
        const pyJson = jsonToPython(parsed, 4);
        args.push(`    json=${pyJson}`);
      } catch {
        args.push(`    data='${req.body!.replace(/'/g, "\\'")}'`);
      }
    } else if (ct?.includes('x-www-form-urlencoded')) {
      // Parse form data into dict
      try {
        const params = new URLSearchParams(req.body!);
        const entries = Array.from(params.entries());
        if (entries.length > 0) {
          const dLines = ['{'];
          for (const [k, v] of entries) {
            dLines.push(`        '${k}': '${v.replace(/'/g, "\\'")}',`);
          }
          dLines.push('    }');
          args.push(`    data=${dLines.join('\n')}`);
        } else {
          args.push(`    data='${req.body!.replace(/'/g, "\\'")}'`);
        }
      } catch {
        args.push(`    data='${req.body!.replace(/'/g, "\\'")}'`);
      }
    } else {
      args.push(`    data='${req.body!.replace(/'/g, "\\'")}'`);
    }
  }

  // Choose helper: requests.get / post / put / patch / delete / request
  const shortMethods = ['get', 'post', 'put', 'patch', 'delete', 'head', 'options'];
  const funcName = shortMethods.includes(methodLower)
    ? `requests.${methodLower}`
    : `requests.request`;

  if (!shortMethods.includes(methodLower)) {
    args.unshift(`    '${req.method}'`);
  }

  lines.push(`response = ${funcName}(`);
  lines.push(args.join(',\n') + ',');
  lines.push(')');

  return lines.join('\n');
}

/** Convert a JS value to Python literal syntax. */
function jsonToPython(value: unknown, indent: number): string {
  const pad = ' '.repeat(indent);
  const innerPad = ' '.repeat(indent + 4);

  if (value === null || value === undefined) return 'None';
  if (typeof value === 'boolean') return value ? 'True' : 'False';
  if (typeof value === 'number') return String(value);
  if (typeof value === 'string') return `'${value.replace(/'/g, "\\'")}'`;

  if (Array.isArray(value)) {
    if (value.length === 0) return '[]';
    const items = value.map((v) => `${innerPad}${jsonToPython(v, indent + 4)}`);
    return `[\n${items.join(',\n')},\n${pad}]`;
  }

  if (typeof value === 'object') {
    const entries = Object.entries(value as Record<string, unknown>);
    if (entries.length === 0) return '{}';
    const items = entries.map(
      ([k, v]) => `${innerPad}'${k}': ${jsonToPython(v, indent + 4)}`
    );
    return `{\n${items.join(',\n')},\n${pad}}`;
  }

  return String(value);
}

// ---------------------------------------------------------------------------
// Go net/http
// ---------------------------------------------------------------------------

export function generateGo(req: CodeGenRequest): string {
  const headers = flatHeaders(req.headers);
  const wantBody = req.body && hasBody(req.method);

  const lines: string[] = [];

  // Body variable
  if (wantBody) {
    lines.push(`body := strings.NewReader(\`${req.body!.replace(/`/g, '` + "`" + `')}\`)`);
    lines.push('');
    lines.push(`req, err := http.NewRequest("${req.method}", "${req.url}", body)`);
  } else {
    lines.push(`req, err := http.NewRequest("${req.method}", "${req.url}", nil)`);
  }

  lines.push('if err != nil {');
  lines.push('    log.Fatal(err)');
  lines.push('}');

  for (const [key, value] of headers) {
    lines.push(`req.Header.Set("${key}", "${value.replace(/"/g, '\\"')}")`);
  }

  lines.push('');
  lines.push('resp, err := http.DefaultClient.Do(req)');
  lines.push('if err != nil {');
  lines.push('    log.Fatal(err)');
  lines.push('}');
  lines.push('defer resp.Body.Close()');

  // Read body
  lines.push('');
  lines.push('respBody, err := io.ReadAll(resp.Body)');
  lines.push('if err != nil {');
  lines.push('    log.Fatal(err)');
  lines.push('}');
  lines.push('fmt.Println(string(respBody))');

  return lines.join('\n');
}
