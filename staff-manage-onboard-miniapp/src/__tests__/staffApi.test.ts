import { describe, it, expect, vi, beforeEach } from 'vitest';
import { staffApi } from '../api/staffApi';

describe('staffApi', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('lists onboarding applications with query parameters', async () => {
    const mockResponse = {
      items: [
        {
          application: {
            id: 'app-uuid-1',
            customer_id: 'CUST-001',
            status: 'REQUIRES_MANUAL_REVIEW',
            current_step: 'kyc_manual_review',
            completed_steps: [],
            personal_info: { first_name: 'John', last_name: 'Doe' },
          },
          step_results: [],
          progress_percent: 60,
          can_submit: false,
        },
      ],
      total: 1,
      limit: 50,
      offset: 0,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    });

    const result = await staffApi.listApplications({
      status: 'REQUIRES_MANUAL_REVIEW',
      search: 'John',
    });

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/onboarding/applications?status=REQUIRES_MANUAL_REVIEW&search=John',
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: 'application/json',
          'Content-Type': 'application/json',
        }),
      })
    );
    expect(result.items.length).toBe(1);
    expect(result.items[0].application.customer_id).toBe('CUST-001');
  });

  it('puts confirmed data to the application', async () => {
    const mockAppDetail = {
      application: {
        id: 'app-uuid-1',
        customer_id: 'CUST-001',
        status: 'REQUIRES_MANUAL_REVIEW',
        confirmed_data: {
          reviewer_id: 'OFFICER-01',
          first_name: 'John Confirmed',
          last_name: 'Doe Confirmed',
          document_number: 'DOC-12345',
          approved_tier: 'TIER_2_PREMIUM',
        },
      },
      step_results: [],
      progress_percent: 60,
      can_submit: false,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockAppDetail,
    });

    const payload = {
      reviewer_id: 'OFFICER-01',
      first_name: 'John Confirmed',
      last_name: 'Doe Confirmed',
      date_of_birth: '1990-01-01',
      nationality: 'VN',
      document_type: 'PASSPORT',
      document_number: 'DOC-12345',
      country: 'VN',
      city: 'Hanoi',
      street_address: '123 Main St',
      postal_code: '100000',
      annual_income: 60000,
      approved_tier: 'TIER_2_PREMIUM',
    };

    const result = await staffApi.putConfirmedData('app-uuid-1', payload);

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/onboarding/app-uuid-1/confirm-data',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify(payload),
      })
    );
    expect(result.application.confirmed_data?.approved_tier).toBe('TIER_2_PREMIUM');
  });

  it('submits manual review decision', async () => {
    const mockResponse = { message: 'Decision submitted' };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    });

    const decisionPayload = {
      reviewer_id: 'OFFICER-01',
      approved: true,
      notes: 'Passed all manual checks',
      confirmed_tier: 'TIER_2_PREMIUM',
    };

    const result = await staffApi.submitManualReviewDecision('app-uuid-1', decisionPayload);

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/onboarding/app-uuid-1/manual-review',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(decisionPayload),
      })
    );
    expect(result.message).toBe('Decision submitted');
  });

  it('handles KYC applications listing and review', async () => {
    const mockKycList = [
      {
        applicationId: 'KYC-2026-001',
        customerId: 'CUST-KYC-1',
        fullName: 'Jane KYC',
        status: 'REQUIRES_MANUAL_REVIEW',
      },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockKycList,
    });

    const list = await staffApi.listKycApplications('REQUIRES_MANUAL_REVIEW');
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/kyc/applications?status=REQUIRES_MANUAL_REVIEW',
      expect.anything()
    );
    expect(list.length).toBe(1);

    // Test KYC review submit
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ message: 'KYC manual review signal submitted successfully' }),
    });

    const reviewRes = await staffApi.submitKycManualReview('KYC-2026-001', {
      reviewerId: 'OFFICER-01',
      decision: 'APPROVE',
      notes: 'Passport and face matched',
    });
    expect(reviewRes.message).toContain('successfully');
  });
});
