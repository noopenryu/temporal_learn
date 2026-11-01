export function generateHex(bytes: number): string {
  const arr = new Uint8Array(bytes);
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(arr);
  } else {
    for (let i = 0; i < bytes; i++) {
      arr[i] = Math.floor(Math.random() * 256);
    }
  }
  return Array.from(arr, (b) => b.toString(16).padStart(2, '0')).join('');
}

export function generateTraceId(): string {
  return generateHex(16); // 32 hex chars
}

export function generateSpanId(): string {
  return generateHex(8); // 16 hex chars
}

const TRACE_ID_KEY = 'temporal_learn_staff_trace_id';

export function getOrCreateTraceId(): string {
  try {
    let id = sessionStorage.getItem(TRACE_ID_KEY);
    if (!id || id.length !== 32) {
      id = generateTraceId();
      sessionStorage.setItem(TRACE_ID_KEY, id);
    }
    return id;
  } catch {
    return generateTraceId();
  }
}

export function resetTraceId(): string {
  const newId = generateTraceId();
  try {
    sessionStorage.setItem(TRACE_ID_KEY, newId);
  } catch {
    // ignore
  }
  return newId;
}

export function getTraceHeaders(): Record<string, string> {
  const traceId = getOrCreateTraceId();
  const spanId = generateSpanId();
  return {
    traceparent: `00-${traceId}-${spanId}-01`,
    'X-Trace-Id': traceId,
    'X-Span-Id': spanId,
  };
}
