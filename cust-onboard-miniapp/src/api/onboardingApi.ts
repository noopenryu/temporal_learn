import {
  AddressInfo,
  ApplicationDetailResponse,
  AuditLog,
  FinancialInfo,
  IdentityDocuments,
  ManualReviewDecision,
  PersonalInfo,
  ResumeSessionRequest,
  StartSessionRequest,
} from '../types/onboarding';
import { getTraceHeaders } from '../observability/tracing';
import { logger } from '../observability/logger';

const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

export class ApiError extends Error {
  status: number;
  data: unknown;
  fieldErrors?: Record<string, string>;

  constructor(message: string, status: number, data?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
    if (data && typeof data === 'object' && 'field_errors' in data) {
      this.fieldErrors = (data as { field_errors: Record<string, string> }).field_errors;
    }
  }
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  const traceHeaders = getTraceHeaders();
  const headers = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    ...traceHeaders,
    ...options.headers,
  };

  const start = performance.now();
  logger.info(`HTTP Request started: ${options.method || 'GET'} ${endpoint}`, {
    endpoint,
    method: options.method || 'GET',
  });

  let response: Response;
  try {
    response = await fetch(url, { ...options, headers });
  } catch (err) {
    const duration = Math.round(performance.now() - start);
    logger.error(`HTTP Request network error: ${options.method || 'GET'} ${endpoint}`, {
      endpoint,
      method: options.method || 'GET',
      duration_ms: duration,
      error: String(err),
    });
    throw err;
  }

  const duration = Math.round(performance.now() - start);
  if (!response.ok) {
    let errorMsg = `HTTP Error ${response.status}: ${response.statusText}`;
    let errorData = null;
    try {
      errorData = await response.json();
      if (errorData && typeof errorData.error === 'string') {
        errorMsg = errorData.error;
      }
    } catch {
      // Body is not JSON
    }
    logger.warn(`HTTP Request failed (${response.status}): ${options.method || 'GET'} ${endpoint}`, {
      endpoint,
      method: options.method || 'GET',
      status_code: response.status,
      duration_ms: duration,
      error: errorMsg,
    });
    throw new ApiError(errorMsg, response.status, errorData);
  }

  logger.info(`HTTP Request completed (${response.status}): ${options.method || 'GET'} ${endpoint}`, {
    endpoint,
    method: options.method || 'GET',
    status_code: response.status,
    duration_ms: duration,
  });

  return response.json();
}

export const onboardingApi = {
  // Start a new session (phone, email, password)
  async startSession(req: StartSessionRequest): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>('/api/v1/client/onboarding/start', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  // Resume an active application using phone + password
  async resumeSession(req: ResumeSessionRequest): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>('/api/v1/client/onboarding/resume', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  // Get application details by ID
  async getApplication(id: string): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}`);
  },

  // Cancel an in-progress draft
  async cancelSession(id: string, reason?: string): Promise<{ message: string }> {
    return request<{ message: string }>(`/api/v1/client/onboarding/${id}/cancel`, {
      method: 'POST',
      body: JSON.stringify({ reason: reason || 'Customer requested cancellation' }),
    });
  },

  // Submit complete application to start Temporal workflow
  async submitApplication(id: string): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/submit`, {
      method: 'POST',
    });
  },

  // Patch personal info step (names, DOB starting from today/18y, gender, phone, email)
  async patchPersonalInfo(id: string, data: Partial<PersonalInfo>): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/personal-info`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Patch address info step
  async patchAddress(id: string, data: Partial<AddressInfo>): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/address`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Patch identity documents step
  async patchIdentityDocs(
    id: string,
    data: Partial<IdentityDocuments>
  ): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/identity-documents`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Patch financial info step
  async patchFinancialInfo(
    id: string,
    data: Partial<FinancialInfo>
  ): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/financial-info`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Patch wallet tier choice
  async patchWalletTier(id: string, walletTier: string): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/wallet-tier`, {
      method: 'PATCH',
      body: JSON.stringify({ wallet_tier: walletTier }),
    });
  },

  // Resubmit identity documents when WAITING_UPDATED_DOCS
  async resubmitDocuments(
    id: string,
    data: Partial<IdentityDocuments>
  ): Promise<{ message: string }> {
    return request<{ message: string }>(`/api/v1/client/onboarding/${id}/resubmit-documents`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // Submit manual review decision (for mock/staff trigger)
  async submitManualReview(
    id: string,
    decision: ManualReviewDecision
  ): Promise<{ message: string }> {
    return request<{ message: string }>(`/api/v1/onboarding/${id}/manual-review`, {
      method: 'POST',
      body: JSON.stringify(decision),
    });
  },

  // Get application audit logs
  async getAuditLogs(id: string): Promise<AuditLog[]> {
    return request<AuditLog[]>(`/api/v1/client/onboarding/${id}/audit-logs`);
  },

  // Get real-time Temporal workflow progress
  async getWorkflowProgress(id: string): Promise<{ stage: string; percent_complete: number; status_message: string; last_updated: string }> {
    return request<{ stage: string; percent_complete: number; status_message: string; last_updated: string }>(
      `/api/v1/client/onboarding/${id}/workflow-progress`
    );
  },

  // Patch whole draft in a single atomic payload
  async patchDraft(id: string, data: Record<string, unknown>): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`/api/v1/client/onboarding/${id}/draft`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  // Upload or simulate document scan upload
  async uploadDocument(req: { doc_type: string; file_name: string; data_url: string }): Promise<{ url: string; quality_score: number; status: string }> {
    return request<{ url: string; quality_score: number; status: string }>(`/api/v1/client/onboarding/upload`, {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  // Health probe
  async checkHealth(): Promise<boolean> {
    try {
      const res = await fetch(`${API_BASE}/healthz`);
      return res.ok;
    } catch {
      return false;
    }
  },
};
