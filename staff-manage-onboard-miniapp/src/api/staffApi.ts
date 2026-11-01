import {
  ApplicationDetailResponse,
  AuditLogsResponse,
  ConfirmedData,
  KycApplicationDto,
  KycManualReviewPayload,
  ListApplicationsResponse,
  ManualReviewDecisionPayload,
} from '../types/staff';
import { getTraceHeaders } from '../observability/tracing';
import { logger } from '../observability/logger';

const ONBOARD_BASE = '/api/v1/onboarding';
const KYC_BASE = '/api/v1/kyc';

class ApiError extends Error {
  status: number;
  data: unknown;

  constructor(message: string, status: number, data?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
  }
}

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const traceHeaders = getTraceHeaders();
  const headers = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    ...traceHeaders,
    ...options.headers,
  };

  const start = performance.now();
  logger.info(`Staff HTTP Request started: ${options.method || 'GET'} ${url}`, {
    url,
    method: options.method || 'GET',
  });

  let response: Response;
  try {
    response = await fetch(url, { ...options, headers });
  } catch (err) {
    const duration = Math.round(performance.now() - start);
    logger.error(`Staff HTTP Request network failure: ${options.method || 'GET'} ${url}`, {
      url,
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
      } else if (errorData && typeof errorData.message === 'string') {
        errorMsg = errorData.message;
      }
    } catch {
      // not JSON
    }
    logger.warn(`Staff HTTP Request failed (${response.status}): ${options.method || 'GET'} ${url}`, {
      url,
      method: options.method || 'GET',
      status_code: response.status,
      duration_ms: duration,
      error: errorMsg,
    });
    throw new ApiError(errorMsg, response.status, errorData);
  }

  logger.info(`Staff HTTP Request completed (${response.status}): ${options.method || 'GET'} ${url}`, {
    url,
    method: options.method || 'GET',
    status_code: response.status,
    duration_ms: duration,
  });

  return response.json();
}

export const staffApi = {
  // --- Customer Onboarding Applications (ms-customer-onboard) ---

  async listApplications(params: {
    status?: string;
    search?: string;
    limit?: number;
    offset?: number;
  } = {}): Promise<ListApplicationsResponse> {
    const query = new URLSearchParams();
    if (params.status && params.status !== 'ALL') query.set('status', params.status);
    if (params.search) query.set('search', params.search);
    if (params.limit) query.set('limit', params.limit.toString());
    if (params.offset) query.set('offset', params.offset.toString());

    const qs = query.toString() ? `?${query.toString()}` : '';
    return request<ListApplicationsResponse>(`${ONBOARD_BASE}/applications${qs}`);
  },

  async getApplication(id: string): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`${ONBOARD_BASE}/${id}`);
  },

  async putConfirmedData(id: string, data: ConfirmedData): Promise<ApplicationDetailResponse> {
    return request<ApplicationDetailResponse>(`${ONBOARD_BASE}/${id}/confirm-data`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  async submitManualReviewDecision(
    id: string,
    payload: ManualReviewDecisionPayload
  ): Promise<{ message: string }> {
    return request<{ message: string }>(`${ONBOARD_BASE}/${id}/manual-review`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async getAuditLogs(id: string): Promise<AuditLogsResponse> {
    return request<AuditLogsResponse>(`${ONBOARD_BASE}/${id}/audit-logs`);
  },

  async getWorkflowProgress(id: string): Promise<{ stage: string; percent_complete: number; status_message: string; last_updated: string }> {
    return request<{ stage: string; percent_complete: number; status_message: string; last_updated: string }>(
      `${ONBOARD_BASE}/${id}/workflow-progress`
    );
  },

  // Helper to generate a demo application that triggers manual review for easy testing
  async seedDemoApplication(scenario: 'review_name' | 'pep' | 'standard' = 'review_name'): Promise<ApplicationDetailResponse> {
    const rand = Math.floor(1000 + Math.random() * 9000);
    const phone = `+84988${rand}${rand.toString().slice(0, 2)}`;
    const email = `applicant.${rand}@walletops-test.internal`;
    const firstName = scenario === 'review_name' ? 'Review Officer Flagged' : 'Tuan';
    const lastName = scenario === 'review_name' ? 'Nguyen Review' : 'Pham';

    // 1. Start Session
    const startRes = await request<ApplicationDetailResponse>(`${ONBOARD_BASE}/start`, {
      method: 'POST',
      body: JSON.stringify({ email, phone_number: phone }),
    });
    const appId = startRes.application.id;

    // 2. Patch Draft in one batch
    await request<ApplicationDetailResponse>(`${ONBOARD_BASE}/${appId}/draft`, {
      method: 'PATCH',
      body: JSON.stringify({
        personal_info: {
          first_name: firstName,
          last_name: lastName,
          date_of_birth: '1992-08-20T00:00:00Z',
          gender: 'MALE',
          nationality: 'VN',
          phone_number: phone,
          email: email,
        },
        address_info: {
          country: 'VN',
          state: 'Hanoi',
          city: 'Hanoi',
          street_address: '88 Ly Thuong Kiet, Hoan Kiem',
          postal_code: '100000',
        },
        identity_docs: {
          document_type: 'NATIONAL_ID',
          document_number: `00109200${rand}`,
          issuing_country: 'VN',
          front_document_url: 'https://images.unsplash.com/photo-1578632767115-351597cf2477?w=600&auto=format&fit=crop&q=80',
          back_document_url: 'https://images.unsplash.com/photo-1589330694653-dad6d3240e2b?w=600&auto=format&fit=crop&q=80',
          selfie_url: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=600&auto=format&fit=crop&q=80',
        },
        financial_info: {
          occupation: 'Financial Analyst',
          employer_name: 'TechFin Global',
          annual_income: scenario === 'pep' ? 120000 : 45000,
          source_of_funds: 'SALARY',
          declared_pep: scenario === 'pep',
          tax_id: `TAX-${rand}`,
        },
        wallet_tier_requested: 'TIER_2_PREMIUM',
      }),
    });

    // 3. Submit application (triggers Temporal workflow)
    const submitRes = await request<ApplicationDetailResponse>(`${ONBOARD_BASE}/${appId}/submit`, {
      method: 'POST',
    });

    return submitRes;
  },

  // --- KYC Microservice (ms-kyc) ---

  async listKycApplications(status?: string): Promise<KycApplicationDto[]> {
    const qs = status && status !== 'ALL' ? `?status=${status}` : '';
    return request<KycApplicationDto[]>(`${KYC_BASE}/applications${qs}`);
  },

  async getKycApplication(id: string): Promise<KycApplicationDto> {
    return request<KycApplicationDto>(`${KYC_BASE}/applications/${id}`);
  },

  async submitKycManualReview(
    applicationId: string,
    payload: KycManualReviewPayload
  ): Promise<{ message: string }> {
    return request<{ message: string }>(`${KYC_BASE}/applications/${applicationId}/manual-review`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async seedDemoKyc(scenario: 'borderline' | 'clear' = 'borderline'): Promise<KycApplicationDto> {
    const rand = Math.floor(1000 + Math.random() * 9000);
    const customerId = `CUST-KYC-${rand}`;
    const payload = {
      customerId,
      fullName: scenario === 'borderline' ? 'Minh Review Tran' : 'Linh Phuong Le',
      dateOfBirth: '1994-04-12',
      nationality: 'VN',
      phoneNumber: `+84912${rand}99`,
      email: `kyc.${rand}@walletops-test.internal`,
      documentType: 'PASSPORT',
      documentNumber: `B${rand}891`,
      frontDocumentUrl: 'https://images.unsplash.com/photo-1544717305-2782549b5136?w=600&auto=format&fit=crop&q=80',
      backDocumentUrl: 'https://images.unsplash.com/photo-1589330694653-dad6d3240e2b?w=600&auto=format&fit=crop&q=80',
      selfieUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=600&auto=format&fit=crop&q=80',
    };

    return request<KycApplicationDto>(`${KYC_BASE}/onboard`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },
};
