export type OnboardingStatus =
  | 'DRAFT'
  | 'SUBMITTED'
  | 'PROCESSING'
  | 'REQUIRES_MANUAL_REVIEW'
  | 'WAITING_UPDATED_DOCS'
  | 'APPROVED'
  | 'REJECTED'
  | 'EXPIRED'
  | 'CANCELLED';

export type WalletTier = 'TIER_1_BASIC' | 'TIER_2_PREMIUM';

export interface PersonalInfo {
  first_name: string;
  last_name: string;
  date_of_birth: string;
  gender?: string;
  nationality: string;
  phone_number: string;
  email: string;
}

export interface AddressInfo {
  country: string;
  state?: string;
  city: string;
  street_address: string;
  postal_code: string;
  proof_of_address_url?: string;
}

export interface IdentityDocuments {
  document_type: string;
  document_number: string;
  issuing_country: string;
  front_document_url: string;
  back_document_url?: string;
  selfie_url: string;
}

export interface FinancialInfo {
  occupation: string;
  employer_name?: string;
  annual_income: number;
  source_of_funds: string;
  declared_pep: boolean;
  tax_id?: string;
}

export interface ConfirmedData {
  confirmed_at?: string;
  reviewer_id: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  nationality: string;
  document_type: string;
  document_number: string;
  country: string;
  city: string;
  street_address: string;
  postal_code: string;
  annual_income: number;
  approved_tier: WalletTier | string;
  review_checklist?: Record<string, boolean>;
  notes?: string;
}

export interface StepCheckResult {
  step_name: string;
  completed: boolean;
  missing_fields?: string[];
}

export interface OnboardingApplication {
  id: string;
  customer_id: string;
  status: OnboardingStatus;
  current_step: string;
  completed_steps: string[];
  personal_info: PersonalInfo;
  address_info: AddressInfo;
  identity_docs: IdentityDocuments;
  financial_info: FinancialInfo;
  confirmed_data?: ConfirmedData;
  wallet_tier_requested: WalletTier | string;
  approved_tier?: WalletTier | string;
  wallet_account_id?: string;
  rejection_reasons?: string[];
  workflow_id?: string;
  run_id?: string;
  expires_at: string;
  submitted_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ApplicationDetailResponse {
  application: OnboardingApplication;
  step_results: StepCheckResult[];
  progress_percent: number;
  can_submit: boolean;
  uncompleted_steps?: string[];
  is_resumed?: boolean;
}

export interface ListApplicationsResponse {
  items: ApplicationDetailResponse[];
  total: number;
  limit: number;
  offset: number;
}

export interface AuditLog {
  id: number;
  application_id: string;
  action: string;
  from_status?: string;
  to_status?: string;
  details?: Record<string, unknown>;
  created_at: string;
}

export interface AuditLogsResponse {
  application_id: string;
  logs: AuditLog[];
  count: number;
}

export interface ManualReviewDecisionPayload {
  reviewer_id: string;
  approved: boolean;
  notes: string;
  confirmed_tier?: string;
  confirmed_data?: ConfirmedData;
}

// ms-kyc types
export type KycStatus =
  | 'PENDING'
  | 'PROCESSING'
  | 'REQUIRES_MANUAL_REVIEW'
  | 'APPROVED'
  | 'REJECTED';

export interface KycApplicationDto {
  applicationId: string;
  customerId: string;
  workflowId: string;
  fullName: string;
  dateOfBirth: string;
  nationality: string;
  phoneNumber: string;
  email: string;
  documentType: string;
  documentNumber: string;
  frontDocumentUrl: string;
  backDocumentUrl?: string;
  selfieUrl: string;
  status: KycStatus;
  walletTier?: string;
  riskScore?: number;
  rejectionReason?: string;
  createdAt?: string;
  updatedAt?: string;
  verificationRecords?: Array<{
    id: number;
    activityName: string;
    passed: boolean;
    score?: number;
    detailsJson?: string;
    createdAt: string;
  }>;
  manualReviews?: Array<{
    id: number;
    reviewerId: string;
    decision: string;
    notes: string;
    reviewedAt: string;
  }>;
}

export interface KycWorkflowProgressDto {
  applicationId: string;
  workflowId: string;
  currentStage: string;
  percentComplete: number;
  statusMessage: string;
}

export interface KycManualReviewPayload {
  reviewerId: string;
  decision: 'APPROVE' | 'REJECT';
  notes: string;
}
