import { generateSpanId, getOrCreateTraceId } from './tracing';

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

interface LogPayload {
  timestamp: string;
  level: LogLevel;
  service: string;
  env: string;
  trace_id: string;
  span_id: string;
  message: string;
  [key: string]: unknown;
}

interface QueuedEntry {
  timestampNs: string;
  line: string;
  level: LogLevel;
}

class FrontendLogger {
  private serviceName: string;
  private env: string;
  private queue: QueuedEntry[] = [];
  private flushTimer: ReturnType<typeof setTimeout> | null = null;
  private isFlushing = false;

  constructor(serviceName: string, env = 'local') {
    this.serviceName = serviceName;
    this.env = env;
  }

  private log(level: LogLevel, message: string, context: Record<string, unknown> = {}) {
    const traceId = getOrCreateTraceId();
    const spanId = generateSpanId();
    const now = new Date();

    const payload: LogPayload = {
      timestamp: now.toISOString(),
      level,
      service: this.serviceName,
      env: this.env,
      trace_id: traceId,
      span_id: spanId,
      message,
      ...context,
    };

    const badgeStyle = 'background: #059669; color: white; padding: 2px 4px; border-radius: 2px; font-weight: bold;';
    const traceStyle = 'color: #64748b; font-size: 11px;';
    switch (level) {
      case 'ERROR':
        console.error(`%c${this.serviceName}%c [${level}] ${message} %c(trace: ${traceId})`, badgeStyle, '', traceStyle, context);
        break;
      case 'WARN':
        console.warn(`%c${this.serviceName}%c [${level}] ${message} %c(trace: ${traceId})`, badgeStyle, '', traceStyle, context);
        break;
      default:
        console.info(`%c${this.serviceName}%c [${level}] ${message} %c(trace: ${traceId})`, badgeStyle, '', traceStyle, context);
        break;
    }

    const timestampNs = (now.getTime() * 1000000).toString();
    this.queue.push({
      timestampNs,
      line: JSON.stringify(payload),
      level,
    });

    if (this.queue.length >= 20) {
      this.flush();
    } else if (!this.flushTimer) {
      this.flushTimer = setTimeout(() => this.flush(), 1000);
    }
  }

  public info(message: string, context?: Record<string, unknown>) {
    this.log('INFO', message, context);
  }

  public warn(message: string, context?: Record<string, unknown>) {
    this.log('WARN', message, context);
  }

  public error(message: string, context?: Record<string, unknown>) {
    this.log('ERROR', message, context);
  }

  public debug(message: string, context?: Record<string, unknown>) {
    this.log('DEBUG', message, context);
  }

  private async flush() {
    if (this.flushTimer) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }

    if (typeof import.meta !== 'undefined' && import.meta.env?.MODE === 'test') {
      this.queue = [];
      return;
    }

    if (this.queue.length === 0 || this.isFlushing) {
      return;
    }

    this.isFlushing = true;
    const batch = this.queue.splice(0, 50);

    const streamsMap: Record<string, { stream: Record<string, string>; values: [string, string][] }> = {};
    for (const item of batch) {
      const key = item.level;
      if (!streamsMap[key]) {
        streamsMap[key] = {
          stream: {
            service: this.serviceName,
            env: this.env,
            level: item.level,
          },
          values: [],
        };
      }
      streamsMap[key].values.push([item.timestampNs, item.line]);
    }

    const payload = {
      streams: Object.values(streamsMap),
    };

    try {
      await fetch('/loki/loki/api/v1/push', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
        keepalive: true,
      });
    } catch {
      // ignore
    } finally {
      this.isFlushing = false;
      if (this.queue.length > 0) {
        this.flushTimer = setTimeout(() => this.flush(), 1000);
      }
    }
  }
}

export const logger = new FrontendLogger('staff-manage-onboard-miniapp');
