import { describe, it, expect, vi, beforeEach } from 'vitest';
import { onboardingApi } from '../api/onboardingApi';

describe('onboardingApi', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('startSession sends POST request to client gateway with password', async () => {
    const mockResponse = {
      application: { id: 'app-uuid-1', customer_id: 'cust-1', status: 'DRAFT' },
      step_results: [],
      progress_percent: 0,
      can_submit: false,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const res = await onboardingApi.startSession({
      email: 'alice@example.com',
      phone_number: '+15551234567',
      password: 'SecretPassword123!',
    });

    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(res.application.id).toBe('app-uuid-1');
    const callArgs = (global.fetch as any).mock.calls[0];
    expect(callArgs[0]).toContain('/api/v1/client/onboarding/start');
    const body = JSON.parse(callArgs[1].body);
    expect(body.password).toBe('SecretPassword123!');
    expect(body.phone_number).toBe('+15551234567');
  });

  it('resumeSession sends POST request with phone and password', async () => {
    const mockResponse = {
      application: { id: 'app-uuid-1', status: 'DRAFT' },
      is_resumed: true,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const res = await onboardingApi.resumeSession({
      phone_number: '+15551234567',
      password: 'SecretPassword123!',
    });

    expect(global.fetch).toHaveBeenCalledTimes(1);
    expect(res.is_resumed).toBe(true);
    const callArgs = (global.fetch as any).mock.calls[0];
    expect(callArgs[0]).toContain('/api/v1/client/onboarding/resume');
    const body = JSON.parse(callArgs[1].body);
    expect(body.phone_number).toBe('+15551234567');
    expect(body.password).toBe('SecretPassword123!');
  });

  it('patchPersonalInfo sends request to client gateway', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ application: { id: 'app-uuid-1' } }),
    } as Response);

    await onboardingApi.patchPersonalInfo('app-uuid-1', {
      first_name: 'Alice',
      last_name: 'Smith',
      date_of_birth: '1995-05-12',
    });

    const callArgs = (global.fetch as any).mock.calls[0];
    expect(callArgs[0]).toContain('/api/v1/client/onboarding/app-uuid-1/personal-info');
    const body = JSON.parse(callArgs[1].body);
    expect(body.first_name).toBe('Alice');
    expect(body.date_of_birth).toBe('1995-05-12');
  });

  it('patchWalletTier sends correct payload', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ application: { id: 'app-uuid-1', wallet_tier_requested: 'TIER_2_PREMIUM' } }),
    } as Response);

    const res = await onboardingApi.patchWalletTier('app-uuid-1', 'TIER_2_PREMIUM');
    const callArgs = (global.fetch as any).mock.calls[0];
    expect(callArgs[0]).toContain('/api/v1/client/onboarding/app-uuid-1/wallet-tier');
    expect(res.application.wallet_tier_requested).toBe('TIER_2_PREMIUM');
  });
});
