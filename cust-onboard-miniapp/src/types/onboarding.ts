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

export type DocumentType = 'NATIONAL_ID' | 'PASSPORT' | 'DRIVING_LICENSE';

export interface PersonalInfo {
  first_name: string;
  last_name: string;
  date_of_birth: string; // ISO 8601 string, e.g. "1995-05-12T00:00:00Z" or "1995-05-12"
  gender?: string;
  nationality: string; // ISO country code e.g. "VN", "SG", "US"
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
  document_type: DocumentType | string;
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
  wallet_tier_requested: string;
  approved_tier?: string;
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

export interface AuditLog {
  id: string;
  application_id: string;
  event_type: string;
  from_status?: string;
  to_status: string;
  reason?: string;
  details?: Record<string, unknown>;
  created_at: string;
}

export interface StartSessionRequest {
  customer_id?: string;
  email: string;
  phone_number: string;
  password?: string;
}

export interface ResumeSessionRequest {
  phone_number: string;
  password: string;
}

export interface ValidationErrorResponse {
  error: string;
  field_errors?: Record<string, string>;
}

export interface ManualReviewDecision {
  reviewer_id: string;
  approved: boolean;
  notes: string;
}

export interface WorkflowProgressResponse {
  stage: string;
  percent_complete: number;
  status_message: string;
  last_updated: string;
}

export interface UploadDocumentResponse {
  url: string;
  quality_score: number;
  status: string;
  created_at: string;
}
