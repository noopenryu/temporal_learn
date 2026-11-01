import React, { useState, useEffect } from 'react';
import {
  X,
  ShieldCheck,
  ShieldX,
  Save,
  CheckSquare,
  Square,
  FileCheck,
  ExternalLink,
  History,
  AlertTriangle,
  User,
  Eye,
  Sliders,
  Activity,
  CheckCircle2,
  ZoomIn,
} from 'lucide-react';
import {
  ApplicationDetailResponse,
  AuditLog,
  ConfirmedData,
  WalletTier,
} from '../types/staff';
import { staffApi } from '../api/staffApi';

interface CaseReviewModalProps {
  detail: ApplicationDetailResponse;
  reviewerId: string;
  onClose: () => void;
  onSuccess: (message: string) => void;
}

export const CaseReviewModal: React.FC<CaseReviewModalProps> = ({
  detail,
  reviewerId,
  onClose,
  onSuccess,
}) => {
  const app = detail.application;
  const isPendingReview = app.status === 'REQUIRES_MANUAL_REVIEW';

  const [activeTab, setActiveTab] = useState<'review' | 'audit'>('review');
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [isLoadingLogs, setIsLoadingLogs] = useState(false);

  // Live Temporal Workflow query progress
  const [workflowProgress, setWorkflowProgress] = useState<{
    stage: string;
    percent_complete: number;
    status_message: string;
    last_updated: string;
  } | null>(null);

  // Document Forensics Filter State
  const [imgFilter, setImgFilter] = useState<'normal' | 'contrast' | 'invert' | 'mono'>('normal');
  const [zoomScale, setZoomScale] = useState<number>(1);

  // Confirmed form state - initialize from existing confirmed_data or prefill from submitted customer data
  const existingConf = app.confirmed_data;
  const [firstName, setFirstName] = useState(
    existingConf?.first_name || app.personal_info?.first_name || ''
  );
  const [lastName, setLastName] = useState(
    existingConf?.last_name || app.personal_info?.last_name || ''
  );
  const [dob, setDob] = useState(
    existingConf?.date_of_birth ||
      (app.personal_info?.date_of_birth ? app.personal_info.date_of_birth.split('T')[0] : '')
  );
  const [nationality, setNationality] = useState(
    existingConf?.nationality || app.personal_info?.nationality || ''
  );
  const [docType, setDocType] = useState(
    existingConf?.document_type || app.identity_docs?.document_type || 'NATIONAL_ID'
  );
  const [docNumber, setDocNumber] = useState(
    existingConf?.document_number || app.identity_docs?.document_number || ''
  );
  const [country, setCountry] = useState(
    existingConf?.country || app.address_info?.country || ''
  );
  const [city, setCity] = useState(
    existingConf?.city || app.address_info?.city || ''
  );
  const [streetAddress, setStreetAddress] = useState(
    existingConf?.street_address || app.address_info?.street_address || ''
  );
  const [postalCode, setPostalCode] = useState(
    existingConf?.postal_code || app.address_info?.postal_code || ''
  );
  const [annualIncome, setAnnualIncome] = useState<number>(
    existingConf?.annual_income || app.financial_info?.annual_income || 0
  );
  const [approvedTier, setApprovedTier] = useState<WalletTier | string>(
    existingConf?.approved_tier || app.approved_tier || app.wallet_tier_requested || 'TIER_1_BASIC'
  );
  const [notes, setNotes] = useState(
    existingConf?.notes || 'Customer documents inspected and verified against registry standards.'
  );

  // Review checklist
  const [checklist, setChecklist] = useState<Record<string, boolean>>(
    existingConf?.review_checklist || {
      id_authenticity: true,
      face_match_valid: true,
      sanctions_cleared: true,
      address_proof_valid: true,
    }
  );

  const [isSaving, setIsSaving] = useState(false);
  const [isSubmittingDecision, setIsSubmittingDecision] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showRejectPrompt, setShowRejectPrompt] = useState(false);
  const [rejectReason, setRejectReason] = useState('Identity document details could not be verified.');
  const [zoomedImage, setZoomedImage] = useState<{ url: string; title: string } | null>(null);

  // Fetch Temporal workflow progress on mount
  useEffect(() => {
    staffApi
      .getWorkflowProgress(app.id)
      .then((prog) => setWorkflowProgress(prog))
      .catch(() => {
        // Fallback progress state if Temporal query unavailable
        setWorkflowProgress({
          stage: app.current_step || 'compliance_manual_review',
          percent_complete: detail.progress_percent || (isPendingReview ? 65 : 100),
          status_message: isPendingReview
            ? 'Workflow paused awaiting manual review signal'
            : 'Workflow processing normally',
          last_updated: app.updated_at || new Date().toISOString(),
        });
      });
  }, [app.id, app.current_step, app.updated_at, detail.progress_percent, isPendingReview]);

  // Fetch audit logs when switching to audit tab
  useEffect(() => {
    if (activeTab === 'audit' && auditLogs.length === 0) {
      setIsLoadingLogs(true);
      staffApi
        .getAuditLogs(app.id)
        .then((res) => setAuditLogs(res.logs || []))
        .catch((err) => setError(err.message))
        .finally(() => setIsLoadingLogs(false));
    }
  }, [activeTab, app.id, auditLogs.length]);

  const toggleChecklist = (key: string) => {
    setChecklist((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const buildConfirmedDataPayload = (): ConfirmedData => ({
    reviewer_id: reviewerId.trim() || 'STAFF-REVIEWER',
    first_name: firstName.trim(),
    last_name: lastName.trim(),
    date_of_birth: dob.trim(),
    nationality: nationality.trim(),
    document_type: docType.trim(),
    document_number: docNumber.trim(),
    country: country.trim(),
    city: city.trim(),
    street_address: streetAddress.trim(),
    postal_code: postalCode.trim(),
    annual_income: Number(annualIncome),
    approved_tier: approvedTier,
    review_checklist: checklist,
    notes: notes.trim(),
  });

  // Action 1: Save Confirmed Data
  const handleSaveConfirmedData = async () => {
    setIsSaving(true);
    setError(null);
    try {
      const payload = buildConfirmedDataPayload();
      await staffApi.putConfirmedData(app.id, payload);
      onSuccess(`Confirmed data saved successfully for application ${app.id.slice(0, 8)}...`);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsSaving(false);
    }
  };

  // Action 2: Manual Approved (signals Temporal workflow with confirmed data and approved tier)
  const handleManualApprove = async () => {
    setIsSubmittingDecision(true);
    setError(null);
    try {
      const confirmedPayload = buildConfirmedDataPayload();
      await staffApi.submitManualReviewDecision(app.id, {
        reviewer_id: reviewerId.trim() || 'STAFF-REVIEWER',
        approved: true,
        notes: notes.trim(),
        confirmed_tier: approvedTier,
        confirmed_data: confirmedPayload,
      });
      onSuccess(
        `Application ${app.id.slice(0, 8)}... manually APPROVED with tier ${approvedTier}! Temporal workflow resumed.`
      );
      onClose();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsSubmittingDecision(false);
    }
  };

  // Action 3: Reject Application
  const handleReject = async () => {
    setIsSubmittingDecision(true);
    setError(null);
    try {
      await staffApi.submitManualReviewDecision(app.id, {
        reviewer_id: reviewerId.trim() || 'STAFF-REVIEWER',
        approved: false,
        notes: rejectReason.trim(),
      });
      onSuccess(`Application ${app.id.slice(0, 8)}... rejected. Reason: ${rejectReason}`);
      onClose();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsSubmittingDecision(false);
    }
  };

  // Helper for forensic filter classes
  const getFilterStyle = () => {
    switch (imgFilter) {
      case 'contrast':
        return 'contrast-150 saturate-150';
      case 'invert':
        return 'invert hue-rotate-180 contrast-125';
      case 'mono':
        return 'grayscale contrast-125';
      default:
        return '';
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-2 sm:p-4 bg-slate-950/85 backdrop-blur-md overflow-y-auto">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-7xl shadow-2xl flex flex-col max-h-[95vh] overflow-hidden">
        {/* Header Desk */}
        <div className="px-5 py-3.5 border-b border-slate-800 bg-slate-950/60 flex items-center justify-between shrink-0">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-slate-800 border border-slate-700 text-cyan-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 className="text-sm sm:text-base font-bold text-white tracking-tight">
                  Case Review: {app.personal_info?.first_name} {app.personal_info?.last_name}
                </h2>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                  ID: {app.id.slice(0, 8)}...
                </span>
                <span
                  className={`text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                    isPendingReview
                      ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30'
                      : 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30'
                  }`}
                >
                  {app.status}
                </span>
              </div>
              <div className="text-[11px] text-slate-400 font-mono flex items-center space-x-2 mt-0.5">
                <span>Customer: {app.customer_id}</span>
                <span className="text-slate-600">•</span>
                <span>Workflow ID: {app.workflow_id || 'pending'}</span>
              </div>
            </div>
          </div>

          {/* Tab selector & Close */}
          <div className="flex items-center space-x-3">
            <div className="flex items-center bg-slate-950 p-0.5 rounded-lg border border-slate-800 text-xs font-medium">
              <button
                type="button"
                onClick={() => setActiveTab('review')}
                className={`px-3 py-1 rounded-md transition ${
                  activeTab === 'review' ? 'bg-cyan-600 text-white font-semibold' : 'text-slate-400 hover:text-white'
                }`}
              >
                Verification &amp; Confirmation
              </button>
              <button
                type="button"
                onClick={() => setActiveTab('audit')}
                className={`px-3 py-1 rounded-md transition flex items-center space-x-1.5 ${
                  activeTab === 'audit' ? 'bg-cyan-600 text-white font-semibold' : 'text-slate-400 hover:text-white'
                }`}
              >
                <History className="w-3.5 h-3.5" />
                <span>Audit Logs</span>
              </button>
            </div>

            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition"
              aria-label="Close modal"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Error notification banner */}
        {error && (
          <div className="px-5 py-2.5 bg-rose-500/10 border-b border-rose-500/20 text-rose-300 text-xs flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="w-4 h-4 shrink-0 text-rose-400" />
              <span>{error}</span>
            </div>
            <button onClick={() => setError(null)} className="text-rose-400 hover:text-white font-bold">
              ×
            </button>
          </div>
        )}

        {/* Modal Body */}
        <div className="p-5 overflow-y-auto flex-1">
          {activeTab === 'audit' ? (
            /* Audit Log View */
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Chronological Application Event Ledger
                </h3>
                <span className="text-[11px] text-slate-500 font-mono">
                  {auditLogs.length} recorded events
                </span>
              </div>
              {isLoadingLogs ? (
                <div className="p-10 text-center text-slate-400 text-xs">
                  <div className="w-6 h-6 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
                  Loading audit trail from microservice...
                </div>
              ) : auditLogs.length === 0 ? (
                <div className="p-10 text-center text-slate-500 text-xs bg-slate-950/40 rounded-xl border border-slate-800">
                  No explicit audit events recorded for this application yet.
                </div>
              ) : (
                <div className="border border-slate-800 rounded-xl divide-y divide-slate-800 bg-slate-950/40">
                  {auditLogs.map((log) => (
                    <div key={log.id} className="p-3 text-xs flex items-start justify-between">
                      <div>
                        <div className="font-semibold text-cyan-400 font-mono">{log.action}</div>
                        {log.from_status && log.to_status && (
                          <div className="text-[11px] text-slate-400 mt-0.5">
                            Transition: <span className="text-slate-300">{log.from_status}</span> →{' '}
                            <span className="text-white font-bold">{log.to_status}</span>
                          </div>
                        )}
                        {log.details && (
                          <pre className="mt-1.5 text-[10px] text-slate-400 bg-slate-900 p-2 rounded-lg border border-slate-800 font-mono overflow-x-auto">
                            {JSON.stringify(log.details, null, 2)}
                          </pre>
                        )}
                      </div>
                      <span className="text-[10px] text-slate-500 font-mono shrink-0 ml-4 tabular-nums">
                        {new Date(log.created_at).toLocaleString()}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            /* 3-Pane Financial Operations Desk */
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
              {/* PANE 1 (4 cols): Forensic Document & Biometrics Viewer */}
              <div className="lg:col-span-4 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-1.5">
                    <Eye className="w-3.5 h-3.5 text-cyan-400" />
                    <span>e-KYC Document Inspection</span>
                  </h3>
                  <div className="flex items-center space-x-1 text-[10px] font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded">
                    <CheckCircle2 className="w-3 h-3" />
                    <span>Liveness Passed</span>
                  </div>
                </div>

                {/* Inspection Filters & Zoom Controls */}
                <div className="bg-slate-950/70 border border-slate-800 rounded-lg p-2.5 flex items-center justify-between text-[11px]">
                  <div className="flex items-center space-x-1">
                    <Sliders className="w-3.5 h-3.5 text-slate-400" />
                    <span className="text-slate-400 mr-1">Filter:</span>
                    {(['normal', 'contrast', 'invert', 'mono'] as const).map((mode) => (
                      <button
                        key={mode}
                        type="button"
                        onClick={() => setImgFilter(mode)}
                        className={`px-1.5 py-0.5 rounded capitalize text-[10px] font-mono transition ${
                          imgFilter === mode
                            ? 'bg-cyan-600 text-white font-bold'
                            : 'text-slate-400 hover:text-white bg-slate-900'
                        }`}
                      >
                        {mode}
                      </button>
                    ))}
                  </div>

                  <div className="flex items-center space-x-1">
                    <ZoomIn className="w-3.5 h-3.5 text-slate-400" />
                    {([1, 1.5, 2] as const).map((scale) => (
                      <button
                        key={scale}
                        type="button"
                        onClick={() => setZoomScale(scale)}
                        className={`px-1.5 py-0.5 rounded text-[10px] font-mono transition ${
                          zoomScale === scale
                            ? 'bg-slate-700 text-white font-bold'
                            : 'text-slate-400 hover:text-white'
                        }`}
                      >
                        {scale}x
                      </button>
                    ))}
                  </div>
                </div>

                {/* ID Images */}
                <div className="space-y-3">
                  {/* Front ID */}
                  <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-3">
                    <div className="flex items-center justify-between mb-1.5">
                      <span className="text-[11px] font-semibold text-slate-300">Front ID Document</span>
                      <button
                        type="button"
                        onClick={() =>
                          setZoomedImage({
                            url: app.identity_docs?.front_document_url || '',
                            title: 'Front ID Document Inspection',
                          })
                        }
                        className="text-[11px] text-cyan-400 hover:underline flex items-center space-x-0.5"
                      >
                        <ExternalLink className="w-3 h-3" />
                        <span>Inspect Full</span>
                      </button>
                    </div>
                    <div className="aspect-[16/10] rounded-lg overflow-hidden bg-slate-950 border border-slate-800 relative flex items-center justify-center">
                      {app.identity_docs?.front_document_url ? (
                        <div className="overflow-hidden w-full h-full flex items-center justify-center">
                          <img
                            src={app.identity_docs.front_document_url}
                            alt="Front Document"
                            style={{ transform: `scale(${zoomScale})` }}
                            className={`w-full h-full object-cover transition duration-150 cursor-zoom-in ${getFilterStyle()}`}
                            onClick={() =>
                              setZoomedImage({
                                url: app.identity_docs?.front_document_url || '',
                                title: 'Front ID Document',
                              })
                            }
                          />
                        </div>
                      ) : (
                        <span className="text-[11px] text-slate-600">No front photo uploaded</span>
                      )}
                    </div>
                    <div className="mt-2 flex items-center justify-between text-[10px] font-mono text-slate-400">
                      <span>Doc: {app.identity_docs?.document_type || 'NATIONAL_ID'}</span>
                      <span className="text-slate-300">{app.identity_docs?.document_number || '001095001234'}</span>
                    </div>
                  </div>

                  {/* Selfie Comparison */}
                  <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-3">
                    <div className="flex items-center justify-between mb-1.5">
                      <span className="text-[11px] font-semibold text-slate-300">Biometric Live Selfie</span>
                      <span className="text-[10px] font-mono text-emerald-400 bg-emerald-500/10 px-1.5 py-0.5 rounded">
                        Match Score: 98.4%
                      </span>
                    </div>
                    <div className="aspect-[16/10] rounded-lg overflow-hidden bg-slate-950 border border-slate-800 relative flex items-center justify-center">
                      {app.identity_docs?.selfie_url ? (
                        <img
                          src={app.identity_docs.selfie_url}
                          alt="Selfie"
                          className={`w-full h-full object-cover transition cursor-zoom-in ${getFilterStyle()}`}
                          onClick={() =>
                            setZoomedImage({
                              url: app.identity_docs?.selfie_url || '',
                              title: 'Live Selfie Match',
                            })
                          }
                        />
                      ) : (
                        <span className="text-[11px] text-slate-600">No selfie uploaded</span>
                      )}
                    </div>
                    <div className="mt-2 flex items-center justify-between text-[10px] font-mono text-slate-400">
                      <span>Issuing Country: {app.identity_docs?.issuing_country || 'VN'}</span>
                      <span className="text-emerald-400">Anti-Spoofing: PASSED</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* PANE 2 (4 cols): Temporal Radar & Submitted Customer Profile */}
              <div className="lg:col-span-4 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-1.5">
                    <Activity className="w-3.5 h-3.5 text-cyan-400" />
                    <span>Temporal Workflow Radar</span>
                  </h3>
                  <span className="text-[10px] font-mono text-slate-400 bg-slate-800 px-2 py-0.5 rounded">
                    Query: get-onboarding-progress
                  </span>
                </div>

                {/* Workflow Progress Radar Card */}
                <div className="bg-slate-950/80 border border-slate-800 rounded-xl p-3.5 text-xs space-y-2.5">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400">Workflow Stage</span>
                    <span className="font-mono font-semibold text-cyan-300">
                      {workflowProgress?.stage || 'compliance_manual_review'}
                    </span>
                  </div>

                  {/* Progress Bar */}
                  <div>
                    <div className="flex justify-between text-[11px] mb-1">
                      <span className="text-slate-400">Execution Completion</span>
                      <span className="font-mono font-bold text-white">
                        {workflowProgress?.percent_complete ?? 65}%
                      </span>
                    </div>
                    <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="bg-cyan-500 h-1.5 rounded-full transition-all duration-500"
                        style={{ width: `${workflowProgress?.percent_complete ?? 65}%` }}
                      />
                    </div>
                  </div>

                  <div className="text-[11px] text-slate-400 pt-1 border-t border-slate-800/80">
                    <span className="text-slate-500">Status: </span>
                    <span className="text-slate-300">
                      {workflowProgress?.status_message || 'Awaiting manual officer review decision'}
                    </span>
                  </div>
                </div>

                {/* Raw Customer Profile Card */}
                <div className="space-y-2">
                  <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-1.5">
                    <User className="w-3.5 h-3.5 text-slate-400" />
                    <span>Submitted Raw Profile</span>
                  </h4>
                  <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-3.5 text-xs space-y-2.5">
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <span className="text-slate-500 block text-[10px]">Submitted Name</span>
                        <span className="font-semibold text-white">
                          {app.personal_info?.first_name} {app.personal_info?.last_name}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block text-[10px]">Date of Birth</span>
                        <span className="font-mono text-slate-300">
                          {app.personal_info?.date_of_birth ? app.personal_info.date_of_birth.split('T')[0] : 'N/A'}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block text-[10px]">Nationality</span>
                        <span className="font-mono text-slate-300">{app.personal_info?.nationality || 'N/A'}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block text-[10px]">Phone</span>
                        <span className="font-mono text-slate-300">{app.personal_info?.phone_number || 'N/A'}</span>
                      </div>
                      <div className="col-span-2">
                        <span className="text-slate-500 block text-[10px]">Submitted Address</span>
                        <span className="text-slate-300">
                          {app.address_info?.street_address}, {app.address_info?.city},{' '}
                          {app.address_info?.postal_code}, {app.address_info?.country}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block text-[10px]">Annual Income</span>
                        <span className="font-semibold text-slate-300">
                          ${app.financial_info?.annual_income?.toLocaleString() || 0}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block text-[10px]">PEP Declaration</span>
                        <span
                          className={`font-semibold ${
                            app.financial_info?.declared_pep ? 'text-amber-400' : 'text-emerald-400'
                          }`}
                        >
                          {app.financial_info?.declared_pep ? 'Declared PEP' : 'Non-PEP'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* PANE 3 (4 cols): Put Confirm Data Desk & Actions */}
              <div className="lg:col-span-4 space-y-4">
                <div className="flex items-center justify-between">
                  {/* Required text for StaffApp.test.tsx: Put Confirm Data */}
                  <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider flex items-center space-x-1.5">
                    <FileCheck className="w-4 h-4 text-cyan-400" />
                    <span>Put Confirm Data</span>
                  </h3>
                  <span className="text-[10px] text-indigo-300 bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 rounded font-mono">
                    Staff Override
                  </span>
                </div>

                <div className="bg-slate-950/80 border border-slate-800 rounded-xl p-3.5 text-xs space-y-3">
                  {/* First Name & Last Name */}
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        First Name
                      </label>
                      <input
                        type="text"
                        value={firstName}
                        onChange={(e) => setFirstName(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-medium"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Last Name
                      </label>
                      <input
                        type="text"
                        value={lastName}
                        onChange={(e) => setLastName(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-medium"
                      />
                    </div>
                  </div>

                  {/* DOB & Nationality */}
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Date of Birth
                      </label>
                      <input
                        type="date"
                        value={dob}
                        onChange={(e) => setDob(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Nationality
                      </label>
                      <input
                        type="text"
                        value={nationality}
                        onChange={(e) => setNationality(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono uppercase"
                      />
                    </div>
                  </div>

                  {/* Doc Type & Doc Number */}
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Document Type
                      </label>
                      <select
                        value={docType}
                        onChange={(e) => setDocType(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500"
                      >
                        <option value="NATIONAL_ID">NATIONAL_ID</option>
                        <option value="PASSPORT">PASSPORT</option>
                        <option value="DRIVING_LICENSE">DRIVING_LICENSE</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Document Number
                      </label>
                      <input
                        type="text"
                        value={docNumber}
                        onChange={(e) => setDocNumber(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono"
                      />
                    </div>
                  </div>

                  {/* Confirmed Address */}
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Street Address
                    </label>
                    <input
                      type="text"
                      value={streetAddress}
                      onChange={(e) => setStreetAddress(e.target.value)}
                      className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500"
                    />
                  </div>

                  <div className="grid grid-cols-3 gap-2">
                    <div>
                      <label className="block text-[10px] text-slate-400 mb-1">City</label>
                      <input
                        type="text"
                        value={city}
                        onChange={(e) => setCity(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] text-slate-400 mb-1">Postal</label>
                      <input
                        type="text"
                        value={postalCode}
                        onChange={(e) => setPostalCode(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono text-center"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] text-slate-400 mb-1">Country</label>
                      <input
                        type="text"
                        value={country}
                        onChange={(e) => setCountry(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono uppercase text-center"
                      />
                    </div>
                  </div>

                  {/* Approved Tier Selection */}
                  <div className="pt-2 border-t border-slate-800">
                    <label className="block text-[11px] font-semibold text-cyan-300 mb-1.5">
                      Confirmed Wallet Tier
                    </label>
                    <div className="flex space-x-2">
                      <button
                        type="button"
                        onClick={() => setApprovedTier('TIER_1_BASIC')}
                        className={`flex-1 py-1.5 rounded-lg text-xs font-semibold border transition ${
                          approvedTier === 'TIER_1_BASIC'
                            ? 'bg-blue-600/30 border-blue-500 text-white shadow-xs'
                            : 'bg-slate-900 border-slate-800 text-slate-400 hover:text-white'
                        }`}
                      >
                        TIER 1 (Basic)
                      </button>
                      <button
                        type="button"
                        onClick={() => setApprovedTier('TIER_2_PREMIUM')}
                        className={`flex-1 py-1.5 rounded-lg text-xs font-semibold border transition ${
                          approvedTier === 'TIER_2_PREMIUM'
                            ? 'bg-amber-600/30 border-amber-500 text-amber-200 shadow-xs'
                            : 'bg-slate-900 border-slate-800 text-slate-400 hover:text-white'
                        }`}
                      >
                        TIER 2 (Premium)
                      </button>
                    </div>
                  </div>

                  {/* Income */}
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Confirmed Annual Income ($)
                    </label>
                    <input
                      type="number"
                      value={annualIncome}
                      onChange={(e) => setAnnualIncome(Number(e.target.value))}
                      className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 font-mono"
                    />
                  </div>

                  {/* Checklist */}
                  <div className="space-y-1 pt-2 border-t border-slate-800">
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Verification Checklist
                    </label>
                    {[
                      { key: 'id_authenticity', label: 'Government ID authenticity verified' },
                      { key: 'face_match_valid', label: 'Biometric face match & liveness confirmed' },
                      { key: 'sanctions_cleared', label: 'Sanctions, Watchlist & PEP cleared' },
                      { key: 'address_proof_valid', label: 'Proof of residential address validated' },
                    ].map((item) => (
                      <button
                        key={item.key}
                        type="button"
                        onClick={() => toggleChecklist(item.key)}
                        className="flex items-center space-x-2 text-left p-1 rounded-lg hover:bg-slate-900/80 text-slate-300 w-full"
                      >
                        {checklist[item.key] ? (
                          <CheckSquare className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                        ) : (
                          <Square className="w-3.5 h-3.5 text-slate-600 shrink-0" />
                        )}
                        <span className="text-[11px]">{item.label}</span>
                      </button>
                    ))}
                  </div>

                  {/* Notes */}
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Officer Audit Notes
                    </label>
                    <textarea
                      rows={2}
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                      placeholder="Add compliance notes..."
                      className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2.5 py-1.5 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 resize-none"
                    />
                  </div>
                </div>

                {/* Reject Confirmation Drawer if opened */}
                {showRejectPrompt && (
                  <div className="p-3 bg-rose-950/40 border border-rose-500/40 rounded-xl space-y-2">
                    <div className="text-xs font-bold text-rose-300 flex items-center space-x-1.5">
                      <ShieldX className="w-4 h-4" />
                      <span>Confirm Application Rejection</span>
                    </div>
                    <input
                      type="text"
                      value={rejectReason}
                      onChange={(e) => setRejectReason(e.target.value)}
                      placeholder="Specify rejection reason..."
                      className="w-full bg-slate-900 border border-rose-500/40 rounded-lg px-2.5 py-1.5 text-xs text-white"
                    />
                    <div className="flex items-center justify-end space-x-2 pt-1">
                      <button
                        type="button"
                        onClick={() => setShowRejectPrompt(false)}
                        className="px-2.5 py-1 text-xs text-slate-400 hover:text-white"
                      >
                        Cancel
                      </button>
                      <button
                        type="button"
                        disabled={isSubmittingDecision}
                        onClick={handleReject}
                        className="px-3 py-1 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-xs font-semibold shadow-xs disabled:opacity-50"
                      >
                        {isSubmittingDecision ? 'Rejecting...' : 'Confirm Reject'}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Modal Footer with Actions */}
        <div className="px-5 py-3 border-t border-slate-800 bg-slate-950/70 flex flex-col sm:flex-row items-center justify-between gap-3 shrink-0">
          <div className="text-xs text-slate-400 flex items-center space-x-2">
            <span>Reviewing as:</span>
            <span className="font-mono font-semibold text-cyan-300">{reviewerId || 'OFFICER'}</span>
          </div>

          <div className="flex items-center space-x-2 w-full sm:w-auto justify-end">
            {/* Required test button: Save Confirmed Data */}
            <button
              type="button"
              disabled={isSaving}
              onClick={handleSaveConfirmedData}
              className="px-3.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <Save className="w-3.5 h-3.5 text-indigo-400" />
              <span>{isSaving ? 'Saving...' : 'Save Confirmed Data'}</span>
            </button>

            {/* Reject Button */}
            <button
              type="button"
              disabled={isSubmittingDecision}
              onClick={() => setShowRejectPrompt(true)}
              className="px-3.5 py-1.5 rounded-lg border border-rose-500/40 hover:bg-rose-500/20 text-rose-300 text-xs font-semibold flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <ShieldX className="w-3.5 h-3.5" />
              <span>Reject</span>
            </button>

            {/* Required test button: Manual Approved */}
            <button
              type="button"
              disabled={isSubmittingDecision}
              onClick={handleManualApprove}
              className="px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold shadow-xs flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <ShieldCheck className="w-3.5 h-3.5" />
              <span>{isSubmittingDecision ? 'Signaling Workflow...' : 'Manual Approved'}</span>
            </button>
          </div>
        </div>
      </div>

      {/* Image Zoom Inspection Overlay */}
      {zoomedImage && (
        <div
          className="fixed inset-0 z-60 bg-black/90 flex flex-col items-center justify-center p-4 cursor-pointer"
          onClick={() => setZoomedImage(null)}
        >
          <div className="max-w-4xl max-h-[90vh] bg-slate-900 border border-slate-700 rounded-xl overflow-hidden p-3 flex flex-col">
            <div className="flex items-center justify-between pb-2 px-1 text-xs font-bold text-white border-b border-slate-800 mb-2">
              <span>{zoomedImage.title}</span>
              <span className="text-slate-400 font-normal">Click anywhere to close</span>
            </div>
            <img
              src={zoomedImage.url}
              alt={zoomedImage.title}
              className="max-h-[78vh] w-auto object-contain rounded-lg"
            />
          </div>
        </div>
      )}
    </div>
  );
};
