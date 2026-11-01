import React, { useState } from 'react';
import {
  ClipboardCheck,
  User,
  MapPin,
  FileText,
  CreditCard,
  Edit2,
  AlertTriangle,
  ArrowLeft,
  XCircle,
  Send,
  CheckCircle2,
} from 'lucide-react';
import { ApplicationDetailResponse } from '../../types/onboarding';

interface ReviewStepProps {
  detail: ApplicationDetailResponse;
  onSubmit: () => Promise<void>;
  onCancel: () => Promise<void>;
  onEditStep: (stepNumber: number) => void;
  onBack: () => void;
  isLoading: boolean;
}

export const ReviewStep: React.FC<ReviewStepProps> = ({
  detail,
  onSubmit,
  onCancel,
  onEditStep,
  onBack,
  isLoading,
}) => {
  const { application, can_submit, uncompleted_steps, progress_percent } = detail;
  const p = application.personal_info;
  const a = application.address_info;
  const i = application.identity_docs;
  const f = application.financial_info;

  const [agreedToTerms, setAgreedToTerms] = useState(true);

  const displayDob =
    p.date_of_birth && !p.date_of_birth.startsWith('0001')
      ? p.date_of_birth.split('T')[0]
      : 'Not specified';

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Title Card */}
      <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs">
        <div className="flex items-center space-x-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
            <ClipboardCheck className="w-5 h-5" aria-hidden="true" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-slate-900 tracking-tight">Review & Confirmation</h2>
            <p className="text-xs text-slate-500">
              Confirm your onboarding details before triggering durable automated saga verification.
            </p>
          </div>
        </div>

        {/* Readiness Checklist Banner */}
        <div
          className={`p-4 rounded-xl border flex items-start space-x-3 ${
            can_submit
              ? 'bg-emerald-50/70 border-emerald-200 text-emerald-900'
              : 'bg-amber-50/70 border-amber-200 text-amber-900'
          }`}
        >
          {can_submit ? (
            <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0 mt-0.5" aria-hidden="true" />
          ) : (
            <AlertTriangle className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" aria-hidden="true" />
          )}
          <div className="flex-1">
            <p className="text-xs font-bold uppercase tracking-wider">
              {can_submit
                ? 'All Steps Complete — Ready for Submission'
                : `Incomplete Application (${progress_percent}% Complete)`}
            </p>
            {!can_submit && uncompleted_steps && uncompleted_steps.length > 0 && (
              <p className="text-xs mt-1 text-amber-800">
                The following sections need your attention:{' '}
                <span className="font-semibold">{uncompleted_steps.join(', ')}</span>.
              </p>
            )}
            {can_submit && (
              <p className="text-xs mt-0.5 text-emerald-700">
                Your application has passed all pre-flight client validation checks and is ready to enter durable execution.
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Breakdown Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* 1. Personal Details */}
        <div className="bg-white border border-slate-200/90 rounded-xl p-5 shadow-2xs relative">
          <div className="flex justify-between items-center mb-3">
            <span className="text-xs font-bold text-slate-800 uppercase tracking-wider flex items-center space-x-1.5">
              <User className="w-4 h-4 text-blue-600" aria-hidden="true" />
              <span>Personal Profile</span>
            </span>
            <button
              type="button"
              onClick={() => onEditStep(1)}
              className="text-xs text-blue-600 hover:text-blue-700 font-semibold flex items-center space-x-1 p-1 rounded"
            >
              <Edit2 className="w-3 h-3" />
              <span>Edit</span>
            </button>
          </div>
          <div className="space-y-1.5 text-xs">
            <div className="flex justify-between">
              <span className="text-slate-500">Legal Name:</span>
              <span className="font-semibold text-slate-900">{p.first_name} {p.last_name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Date of Birth:</span>
              <span className="font-mono text-slate-900">{displayDob}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Nationality:</span>
              <span className="font-semibold text-slate-900">{p.nationality}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Phone:</span>
              <span className="font-mono text-slate-900">{p.phone_number}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Email:</span>
              <span className="text-slate-900">{p.email}</span>
            </div>
          </div>
        </div>

        {/* 2. Address Details */}
        <div className="bg-white border border-slate-200/90 rounded-xl p-5 shadow-2xs relative">
          <div className="flex justify-between items-center mb-3">
            <span className="text-xs font-bold text-slate-800 uppercase tracking-wider flex items-center space-x-1.5">
              <MapPin className="w-4 h-4 text-blue-600" aria-hidden="true" />
              <span>Residency & Address</span>
            </span>
            <button
              type="button"
              onClick={() => onEditStep(2)}
              className="text-xs text-blue-600 hover:text-blue-700 font-semibold flex items-center space-x-1 p-1 rounded"
            >
              <Edit2 className="w-3 h-3" />
              <span>Edit</span>
            </button>
          </div>
          <div className="space-y-1.5 text-xs">
            <div className="flex justify-between">
              <span className="text-slate-500">Country:</span>
              <span className="font-semibold text-slate-900">{a.country}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">City / State:</span>
              <span className="text-slate-900">{a.city}{a.state ? `, ${a.state}` : ''}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Postal Code:</span>
              <span className="font-mono text-slate-900">{a.postal_code}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Street:</span>
              <span className="text-slate-900 truncate max-w-[180px]">{a.street_address}</span>
            </div>
          </div>
        </div>

        {/* 3. Identity & Documents */}
        <div className="bg-white border border-slate-200/90 rounded-xl p-5 shadow-2xs relative">
          <div className="flex justify-between items-center mb-3">
            <span className="text-xs font-bold text-slate-800 uppercase tracking-wider flex items-center space-x-1.5">
              <FileText className="w-4 h-4 text-blue-600" aria-hidden="true" />
              <span>Identity (e-KYC)</span>
            </span>
            <button
              type="button"
              onClick={() => onEditStep(3)}
              className="text-xs text-blue-600 hover:text-blue-700 font-semibold flex items-center space-x-1 p-1 rounded"
            >
              <Edit2 className="w-3 h-3" />
              <span>Edit</span>
            </button>
          </div>
          <div className="space-y-1.5 text-xs">
            <div className="flex justify-between">
              <span className="text-slate-500">Document Type:</span>
              <span className="font-semibold text-slate-900">{i.document_type}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Doc Number:</span>
              <span className="font-mono font-semibold text-slate-900">{i.document_number}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Issuing Country:</span>
              <span className="font-semibold text-slate-900">{i.issuing_country}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Biometric Selfie:</span>
              <span className="text-emerald-700 font-semibold flex items-center">
                <CheckCircle2 className="w-3.5 h-3.5 mr-1 text-emerald-600" /> Attached
              </span>
            </div>
          </div>
        </div>

        {/* 4. Financial & Tier */}
        <div className="bg-white border border-slate-200/90 rounded-xl p-5 shadow-2xs relative">
          <div className="flex justify-between items-center mb-3">
            <span className="text-xs font-bold text-slate-800 uppercase tracking-wider flex items-center space-x-1.5">
              <CreditCard className="w-4 h-4 text-blue-600" aria-hidden="true" />
              <span>Financial & Tier</span>
            </span>
            <button
              type="button"
              onClick={() => onEditStep(5)}
              className="text-xs text-blue-600 hover:text-blue-700 font-semibold flex items-center space-x-1 p-1 rounded"
            >
              <Edit2 className="w-3 h-3" />
              <span>Edit</span>
            </button>
          </div>
          <div className="space-y-1.5 text-xs">
            <div className="flex justify-between">
              <span className="text-slate-500">Requested Tier:</span>
              <span className="font-bold text-blue-700">
                {application.wallet_tier_requested === 'TIER_2_PREMIUM' ? 'Tier 2: Premium Wallet' : 'Tier 1: Basic Wallet'}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Occupation:</span>
              <span className="text-slate-900">{f.occupation}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Annual Income:</span>
              <span className="font-mono font-semibold text-slate-900 tabular-nums">
                ${f.annual_income?.toLocaleString()} USD
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">PEP Declared:</span>
              <span className={f.declared_pep ? 'text-amber-700 font-bold' : 'text-slate-900'}>
                {f.declared_pep ? 'YES' : 'NO'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Terms & Regulatory Certification */}
      <div className="bg-white border border-slate-200/90 rounded-xl p-4 shadow-2xs">
        <label className="flex items-start space-x-3 cursor-pointer">
          <input
            type="checkbox"
            checked={agreedToTerms}
            onChange={(e) => setAgreedToTerms(e.target.checked)}
            className="mt-0.5 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
          />
          <span className="text-xs text-slate-600 leading-relaxed">
            I certify under penalty of perjury that all identity, tax, and employment information provided is authentic and correct. I authorize PayFlow to perform automated KYC/AML checks and execute durable onboarding workflows.
          </span>
        </label>
      </div>

      {/* Submission Actions */}
      <div className="bg-white border border-slate-200/90 rounded-2xl p-5 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-xs">
        <div className="flex items-center space-x-3 w-full sm:w-auto">
          <button
            type="button"
            onClick={onBack}
            className="flex items-center space-x-2 px-4 py-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 font-semibold text-xs transition-colors"
          >
            <ArrowLeft className="w-4 h-4" aria-hidden="true" />
            <span>Back</span>
          </button>

          <button
            type="button"
            onClick={onCancel}
            disabled={isLoading}
            className="flex items-center space-x-2 px-4 py-2.5 rounded-xl border border-rose-200 text-rose-700 hover:bg-rose-50 font-semibold text-xs transition-colors disabled:opacity-50"
          >
            <XCircle className="w-4 h-4" aria-hidden="true" />
            <span>Abandon Draft</span>
          </button>
        </div>

        <button
          type="button"
          disabled={!can_submit || !agreedToTerms || isLoading}
          onClick={onSubmit}
          className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-8 rounded-xl shadow-md shadow-blue-500/25 flex items-center justify-center space-x-2 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <Send className="w-4 h-4" aria-hidden="true" />
          <span>{isLoading ? 'Submitting to Temporal...' : 'Submit Application'}</span>
        </button>
      </div>
    </div>
  );
};
