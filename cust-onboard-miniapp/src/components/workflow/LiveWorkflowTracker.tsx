import React, { useEffect, useState } from 'react';
import {
  CheckCircle2,
  Clock,
  FileWarning,
  ShieldAlert,
  XCircle,
  History,
  Activity,
} from 'lucide-react';
import {
  ApplicationDetailResponse,
  IdentityDocuments,
  ManualReviewDecision,
  OnboardingStatus,
  WorkflowProgressResponse,
} from '../../types/onboarding';
import { onboardingApi } from '../../api/onboardingApi';
import { DocumentResubmissionModal } from './DocumentResubmissionModal';
import { ComplianceReviewPanel } from './ComplianceReviewPanel';
import { WalletSuccessCard } from './WalletSuccessCard';
import { AuditLogsDrawer } from './AuditLogsDrawer';

interface LiveWorkflowTrackerProps {
  initialDetail: ApplicationDetailResponse;
  onNewSession: () => void;
}

export const LiveWorkflowTracker: React.FC<LiveWorkflowTrackerProps> = ({
  initialDetail,
  onNewSession,
}) => {
  const [detail, setDetail] = useState<ApplicationDetailResponse>(initialDetail);
  const [workflowProgress, setWorkflowProgress] = useState<WorkflowProgressResponse | null>(null);
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [auditLogsOpen, setAuditLogsOpen] = useState(false);

  const { application } = detail;
  const status = application.status;

  // Poll backend every 2s if workflow is in-flight
  useEffect(() => {
    const activeStatuses: OnboardingStatus[] = [
      'SUBMITTED',
      'PROCESSING',
      'WAITING_UPDATED_DOCS',
      'REQUIRES_MANUAL_REVIEW',
    ];

    if (!activeStatuses.includes(status)) {
      return;
    }

    const poll = async () => {
      try {
        const [updated, progress] = await Promise.all([
          onboardingApi.getApplication(application.id),
          onboardingApi.getWorkflowProgress(application.id).catch(() => null),
        ]);
        setDetail(updated);
        if (progress) setWorkflowProgress(progress);
      } catch (err) {
        console.error('Polling error:', err);
      }
    };

    poll();
    const interval = setInterval(poll, 2000);
    return () => clearInterval(interval);
  }, [application.id, status]);

  // Handle document resubmission signal
  const handleResubmitDocs = async (docs: Partial<IdentityDocuments>) => {
    setIsActionLoading(true);
    try {
      await onboardingApi.resubmitDocuments(application.id, docs);
      const updated = await onboardingApi.getApplication(application.id);
      setDetail(updated);
    } catch (err) {
      console.error('Resubmit error:', err);
      alert('Failed to resubmit documents: ' + (err as Error).message);
    } finally {
      setIsActionLoading(false);
    }
  };

  // Handle manual compliance decision signal
  const handleManualReview = async (decision: ManualReviewDecision) => {
    setIsActionLoading(true);
    try {
      await onboardingApi.submitManualReview(application.id, decision);
      const updated = await onboardingApi.getApplication(application.id);
      setDetail(updated);
    } catch (err) {
      console.error('Review decision error:', err);
      alert('Failed to submit manual review: ' + (err as Error).message);
    } finally {
      setIsActionLoading(false);
    }
  };

  // Render finalized card if Approved
  if (status === 'APPROVED') {
    return (
      <div className="space-y-6">
        <WalletSuccessCard application={application} onNewSession={onNewSession} />
        <div className="flex justify-center">
          <button
            type="button"
            onClick={() => setAuditLogsOpen(true)}
            className="text-xs text-slate-500 hover:text-blue-600 flex items-center space-x-1.5 transition font-semibold"
          >
            <History className="w-3.5 h-3.5" />
            <span>View Full Workflow Audit Trail</span>
          </button>
        </div>
        <AuditLogsDrawer
          applicationId={application.id}
          isOpen={auditLogsOpen}
          onClose={() => setAuditLogsOpen(false)}
        />
      </div>
    );
  }

  // Calculate visual pipeline activities
  const getStageStatus = (stage: 'submitted' | 'eligibility' | 'kyc' | 'provisioning') => {
    if (status === 'REJECTED') return 'rejected';

    switch (stage) {
      case 'submitted':
        return 'completed';
      case 'eligibility':
        if (status === 'SUBMITTED') return 'active';
        return 'completed';
      case 'kyc':
        if (status === 'PROCESSING') return 'active';
        if (status === 'WAITING_UPDATED_DOCS' || status === 'REQUIRES_MANUAL_REVIEW') return 'paused';
        return 'pending';
      case 'provisioning':
        return 'pending';
    }
  };

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Top Banner: Real-Time Workflow Status */}
      <div className="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-xs space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-100">
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                Temporal Durable Saga Orchestration
              </span>
              <span className="w-1.5 h-1.5 rounded-full bg-slate-300" />
              <span className="text-xs font-mono text-slate-600">App ID: {application.id.slice(0, 8)}...</span>
            </div>

            <h2 className="text-xl sm:text-2xl font-bold text-slate-900 flex flex-wrap items-center gap-2.5">
              <span>Onboarding Saga Pipeline</span>
              <span
                className={`text-xs font-extrabold px-3 py-1 rounded-full border uppercase tracking-wider flex items-center space-x-1.5 ${
                  status === 'PROCESSING' || status === 'SUBMITTED'
                    ? 'bg-blue-50 text-blue-700 border-blue-200'
                    : status === 'WAITING_UPDATED_DOCS'
                    ? 'bg-amber-50 text-amber-800 border-amber-300'
                    : status === 'REQUIRES_MANUAL_REVIEW'
                    ? 'bg-purple-50 text-purple-700 border-purple-200'
                    : status === 'REJECTED'
                    ? 'bg-rose-50 text-rose-700 border-rose-200'
                    : 'bg-emerald-50 text-emerald-700 border-emerald-200'
                }`}
              >
                {(status === 'PROCESSING' || status === 'SUBMITTED') && (
                  <span className="w-2 h-2 rounded-full bg-blue-600 animate-pulse" />
                )}
                <span>{status.replace(/_/g, ' ')}</span>
              </span>
            </h2>
          </div>

          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={() => setAuditLogsOpen(true)}
              className="px-3 py-2 rounded-xl border border-slate-200 hover:bg-slate-50 text-slate-700 text-xs font-semibold flex items-center space-x-1.5 transition shadow-2xs"
            >
              <History className="w-4 h-4 text-slate-500" />
              <span>Audit Trail</span>
            </button>
          </div>
        </div>

        {/* Real-time Temporal Query Progress Bar */}
        {workflowProgress && (
          <div className="p-4 rounded-xl bg-blue-50/60 border border-blue-200/80">
            <div className="flex justify-between items-center text-xs font-semibold text-slate-800 mb-1.5">
              <span className="flex items-center space-x-1.5">
                <Activity className="w-3.5 h-3.5 text-blue-600 animate-spin" />
                <span>Stage: {workflowProgress.stage}</span>
              </span>
              <span className="font-mono text-blue-700 text-sm font-bold">{workflowProgress.percent_complete}%</span>
            </div>
            <div className="w-full bg-blue-100 rounded-full h-2 overflow-hidden mb-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-500 ease-out"
                style={{ width: `${workflowProgress.percent_complete}%` }}
              />
            </div>
            <p className="text-xs text-slate-600">{workflowProgress.status_message}</p>
          </div>
        )}

        {/* Workflow Info Chips */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
          <div className="p-3 rounded-xl bg-slate-50 border border-slate-200/60">
            <span className="text-slate-500 block text-[10px] font-semibold uppercase">Verified Contact</span>
            <span className="font-mono font-medium text-slate-900 mt-0.5 block">{application.personal_info?.phone_number || 'N/A'}</span>
          </div>

          <div className="p-3 rounded-xl bg-slate-50 border border-slate-200/60">
            <span className="text-slate-500 block text-[10px] font-semibold uppercase">Temporal Workflow ID</span>
            <span className="font-mono text-[11px] text-blue-700 truncate mt-0.5 block" title={application.workflow_id}>
              {application.workflow_id || 'Starting in Temporal...'}
            </span>
          </div>

          <div className="p-3 rounded-xl bg-slate-50 border border-slate-200/60">
            <span className="text-slate-500 block text-[10px] font-semibold uppercase">Current Execution Activity</span>
            <span className="font-mono text-slate-900 mt-0.5 block capitalize">
              {application.current_step.replace(/_/g, ' ')}
            </span>
          </div>
        </div>

        {/* Visual Workflow Activities Pipeline */}
        <div className="space-y-3 pt-2">
          <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider">
            Multi-Service Orchestration Pipeline
          </h4>

          <div className="space-y-2.5">
            {/* Step 1: Submissions */}
            <div className="flex items-center space-x-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/70">
              <div className="w-7 h-7 rounded-lg bg-emerald-100 text-emerald-700 flex items-center justify-center shrink-0">
                <CheckCircle2 className="w-4 h-4" />
              </div>
              <div className="flex-1">
                <p className="text-xs font-bold text-slate-900">Application Dispatch & Schema Validation</p>
                <p className="text-[11px] text-slate-500">Temporal Workflow started with saga compensation checkpoints</p>
              </div>
              <span className="text-[10px] font-bold font-mono px-2 py-0.5 rounded bg-emerald-100 text-emerald-800">PASSED</span>
            </div>

            {/* Step 2: Eligibility Check */}
            <div className="flex items-center space-x-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/70">
              <div
                className={`w-7 h-7 rounded-lg flex items-center justify-center shrink-0 ${
                  getStageStatus('eligibility') === 'completed'
                    ? 'bg-emerald-100 text-emerald-700'
                    : 'bg-blue-100 text-blue-700'
                }`}
              >
                {getStageStatus('eligibility') === 'completed' ? (
                  <CheckCircle2 className="w-4 h-4" />
                ) : (
                  <Clock className="w-4 h-4" />
                )}
              </div>
              <div className="flex-1">
                <p className="text-xs font-bold text-slate-900">
                  Regulatory & Eligibility Rules Scoring (<code className="text-[11px] text-blue-700 font-mono">ms-eligiblity</code>)
                </p>
                <p className="text-[11px] text-slate-500">FATF sanctions, PEP screening, age validation & credit risk score</p>
              </div>
              <span className="text-[10px] font-bold font-mono px-2 py-0.5 rounded bg-slate-100 text-slate-700">
                {getStageStatus('eligibility') === 'completed' ? 'PASSED' : 'EVALUATING'}
              </span>
            </div>

            {/* Step 3: KYC Verification */}
            <div className="flex items-center space-x-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/70">
              <div
                className={`w-7 h-7 rounded-lg flex items-center justify-center shrink-0 ${
                  status === 'WAITING_UPDATED_DOCS'
                    ? 'bg-amber-100 text-amber-700'
                    : status === 'REQUIRES_MANUAL_REVIEW'
                    ? 'bg-purple-100 text-purple-700'
                    : getStageStatus('kyc') === 'active'
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-slate-200 text-slate-500'
                }`}
              >
                {status === 'WAITING_UPDATED_DOCS' ? (
                  <FileWarning className="w-4 h-4" />
                ) : status === 'REQUIRES_MANUAL_REVIEW' ? (
                  <ShieldAlert className="w-4 h-4" />
                ) : (
                  <Clock className="w-4 h-4" />
                )}
              </div>
              <div className="flex-1">
                <p className="text-xs font-bold text-slate-900">
                  Identity & Biometric KYC (<code className="text-[11px] text-blue-700 font-mono">ms-kyc</code>)
                </p>
                <p className="text-[11px] text-slate-500">
                  Document OCR, facial selfie verification, and fraud detection
                </p>
              </div>
              <span className="text-[10px] font-bold font-mono px-2 py-0.5 rounded bg-slate-100 text-slate-700">
                {status === 'WAITING_UPDATED_DOCS'
                  ? 'WAITING DOCS'
                  : status === 'REQUIRES_MANUAL_REVIEW'
                  ? 'PENDING SIGN-OFF'
                  : 'PROCESSING'}
              </span>
            </div>

            {/* Step 4: Digital Wallet Account Provisioning */}
            <div className="flex items-center space-x-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/70">
              <div className="w-7 h-7 rounded-lg bg-slate-200 text-slate-500 flex items-center justify-center shrink-0">
                <Clock className="w-4 h-4" />
              </div>
              <div className="flex-1">
                <p className="text-xs font-bold text-slate-500">Digital Wallet Account Ledger Provisioning</p>
                <p className="text-[11px] text-slate-400">Generating unique account ID & setting tier balance limits</p>
              </div>
              <span className="text-[10px] font-bold font-mono px-2 py-0.5 rounded bg-slate-100 text-slate-500">
                PENDING
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Dynamic Resolution: Document Resubmission Signal Trigger */}
      {status === 'WAITING_UPDATED_DOCS' && (
        <DocumentResubmissionModal
          initialDocs={application.identity_docs}
          rejectionReason={
            application.rejection_reasons?.[0] ||
            'Document image was flagged as blurry, poorly lit, or unreadable.'
          }
          onResubmit={handleResubmitDocs}
          isLoading={isActionLoading}
        />
      )}

      {/* Dynamic Resolution: Compliance Manual Review Signal Trigger */}
      {status === 'REQUIRES_MANUAL_REVIEW' && (
        <ComplianceReviewPanel onDecision={handleManualReview} isLoading={isActionLoading} />
      )}

      {/* Rejected State */}
      {status === 'REJECTED' && (
        <div className="bg-white border-2 border-rose-300 rounded-2xl p-6 shadow-xs">
          <div className="flex items-start space-x-3">
            <XCircle className="w-6 h-6 text-rose-600 shrink-0 mt-0.5" />
            <div className="flex-1">
              <h3 className="text-lg font-bold text-slate-900">Application Unsuccessful</h3>
              <p className="text-xs text-rose-700 mt-1">
                Your customer onboarding application was evaluated and could not be approved at this time.
              </p>
              {application.rejection_reasons && application.rejection_reasons.length > 0 && (
                <div className="mt-3 p-3.5 rounded-xl bg-rose-50 border border-rose-200">
                  <p className="text-xs font-bold text-rose-900 mb-1">Rejection Details:</p>
                  <ul className="list-disc list-inside text-xs text-rose-800 space-y-1">
                    {application.rejection_reasons.map((reason, idx) => (
                      <li key={idx}>{reason}</li>
                    ))}
                  </ul>
                </div>
              )}
              <div className="mt-5">
                <button
                  type="button"
                  onClick={onNewSession}
                  className="px-5 py-2.5 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-semibold text-xs transition shadow-xs"
                >
                  Start New Application
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Audit Logs Drawer */}
      <AuditLogsDrawer
        applicationId={application.id}
        isOpen={auditLogsOpen}
        onClose={() => setAuditLogsOpen(false)}
      />
    </div>
  );
};
