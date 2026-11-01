import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { App } from '../App';
import { staffApi } from '../api/staffApi';

const mockApplicationDetail = {
  application: {
    id: 'f87a3b21-1234-4567-8901-abcdef123456',
    customer_id: 'CUST-TEST-100',
    status: 'REQUIRES_MANUAL_REVIEW' as const,
    current_step: 'compliance_manual_review',
    completed_steps: ['personal_info', 'address_info', 'identity_documents'],
    personal_info: {
      first_name: 'Minh',
      last_name: 'Review Test',
      date_of_birth: '1995-10-15T00:00:00Z',
      gender: 'MALE',
      nationality: 'VN',
      phone_number: '+84988776655',
      email: 'minh.test@example.com',
    },
    address_info: {
      country: 'VN',
      state: 'Hanoi',
      city: 'Hanoi',
      street_address: '45 Trang Thi',
      postal_code: '100000',
    },
    identity_docs: {
      document_type: 'NATIONAL_ID',
      document_number: '001095001234',
      issuing_country: 'VN',
      front_document_url: 'https://example.com/front.jpg',
      selfie_url: 'https://example.com/selfie.jpg',
    },
    financial_info: {
      occupation: 'Software Engineer',
      annual_income: 50000,
      source_of_funds: 'SALARY',
      declared_pep: false,
    },
    wallet_tier_requested: 'TIER_1_BASIC' as const,
    workflow_id: 'onboard-wf-test-1',
    expires_at: '2026-11-15T00:00:00Z',
    created_at: '2026-10-30T10:00:00Z',
    updated_at: '2026-10-30T10:05:00Z',
  },
  step_results: [],
  progress_percent: 65,
  can_submit: false,
};

describe('Staff Workbench App', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders workbench title and queues correctly', async () => {
    vi.spyOn(staffApi, 'listApplications').mockResolvedValue({
      items: [mockApplicationDetail],
      total: 1,
      limit: 50,
      offset: 0,
    });

    render(<App />);

    expect(screen.getByText('WalletOps Workbench')).toBeInTheDocument();
    expect(screen.getByText('Staff Review Portal')).toBeInTheDocument();
    expect(screen.getByText('Customer Onboarding Queue')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Minh Review Test')).toBeInTheDocument();
      expect(screen.getByText('CUST-TEST-100')).toBeInTheDocument();
      expect(screen.getByText('Manual Review Required')).toBeInTheDocument();
    });
  });

  it('opens review modal, allows editing confirmed data, and manual approval', async () => {
    vi.spyOn(staffApi, 'listApplications').mockResolvedValue({
      items: [mockApplicationDetail],
      total: 1,
      limit: 50,
      offset: 0,
    });

    const putSpy = vi.spyOn(staffApi, 'putConfirmedData').mockResolvedValue(mockApplicationDetail);
    const decisionSpy = vi.spyOn(staffApi, 'submitManualReviewDecision').mockResolvedValue({
      message: 'Decision recorded',
    });

    render(<App />);

    // Click "Review & Confirm"
    const reviewBtn = await screen.findByText('Review & Confirm');
    fireEvent.click(reviewBtn);

    // Modal opens
    await waitFor(() => {
      expect(screen.getByText(/Put Confirm Data/i)).toBeInTheDocument();
      expect(screen.getByDisplayValue('Minh')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Review Test')).toBeInTheDocument();
      expect(screen.getByDisplayValue('001095001234')).toBeInTheDocument();
    });

    // Test Save Confirmed Data
    const saveConfirmedBtn = screen.getByText('Save Confirmed Data');
    fireEvent.click(saveConfirmedBtn);

    await waitFor(() => {
      expect(putSpy).toHaveBeenCalled();
    });

    // Test Manual Approved
    const manualApproveBtn = screen.getByText('Manual Approved');
    fireEvent.click(manualApproveBtn);

    await waitFor(() => {
      expect(decisionSpy).toHaveBeenCalledWith(
        'f87a3b21-1234-4567-8901-abcdef123456',
        expect.objectContaining({
          approved: true,
        })
      );
    });
  });

  it('switches to KYC queue when KYC tab is clicked', async () => {
    vi.spyOn(staffApi, 'listApplications').mockResolvedValue({
      items: [],
      total: 0,
      limit: 50,
      offset: 0,
    });

    const listKycSpy = vi.spyOn(staffApi, 'listKycApplications').mockResolvedValue([
      {
        applicationId: 'KYC-2026-999',
        customerId: 'CUST-KYC-999',
        workflowId: 'kyc-wf-999',
        fullName: 'Le Van KYC',
        dateOfBirth: '1988-12-01',
        nationality: 'VN',
        phoneNumber: '+84900000001',
        email: 'kyc@example.com',
        documentType: 'PASSPORT',
        documentNumber: 'P1234567',
        frontDocumentUrl: 'https://example.com/p.jpg',
        selfieUrl: 'https://example.com/s.jpg',
        status: 'REQUIRES_MANUAL_REVIEW',
        riskScore: 65,
      },
    ]);

    render(<App />);

    const kycTab = screen.getByText('KYC Microservice Cases');
    fireEvent.click(kycTab);

    await waitFor(() => {
      expect(listKycSpy).toHaveBeenCalled();
      expect(screen.getByText('Le Van KYC')).toBeInTheDocument();
      expect(screen.getByText('Review KYC')).toBeInTheDocument();
    });
  });
});
