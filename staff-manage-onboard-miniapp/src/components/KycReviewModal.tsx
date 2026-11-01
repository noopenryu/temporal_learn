import React, { useState, useEffect } from 'react';
import {
  X,
  ShieldCheck,
  ShieldX,
  AlertTriangle,
  Eye,
  Activity,
} from 'lucide-react';
import { KycApplicationDto } from '../types/staff';
import { staffApi } from '../api/staffApi';

interface KycReviewModalProps {
  kyc: KycApplicationDto;
  reviewerId: string;
  onClose: () => void;
  onSuccess: (message: string) => void;
}

export const KycReviewModal: React.FC<KycReviewModalProps> = ({
  kyc,
  reviewerId,
  onClose,
  onSuccess,
}) => {
  const [notes, setNotes] = useState(
    'Biometrics and ID document verified manually against sanctions and internal watchlists.'
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [zoomedImage, setZoomedImage] = useState<{ url: string; title: string } | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (zoomedImage) {
          setZoomedImage(null);
        } else {
          onClose();
        }
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose, zoomedImage]);

  const handleAction = async (decision: 'APPROVE' | 'REJECT') => {
    setIsSubmitting(true);
    setError(null);
    try {
      await staffApi.submitKycManualReview(kyc.applicationId, {
        reviewerId: reviewerId.trim() || 'OFFICER-KYC-01',
        decision,
        notes: notes.trim(),
      });
      onSuccess(
        `KYC application ${kyc.applicationId} ${
          decision === 'APPROVE' ? 'APPROVED' : 'REJECTED'
        }! Temporal KYC workflow resumed.`
      );
      onClose();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-5 bg-slate-950/85 backdrop-blur-md overflow-y-auto">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-4xl shadow-2xl flex flex-col max-h-[92vh] overflow-hidden">
        {/* Header */}
        <div className="px-5 py-3.5 border-b border-slate-800 bg-slate-950/60 flex items-center justify-between shrink-0">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-slate-800 border border-slate-700 text-cyan-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 className="text-sm sm:text-base font-bold text-white tracking-tight">KYC Review: {kyc.fullName}</h2>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                  {kyc.applicationId}
                </span>
                <span className="text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-300 border border-amber-500/30">
                  {kyc.status}
                </span>
              </div>
              <div className="text-[11px] text-slate-400 font-mono mt-0.5">
                Customer: {kyc.customerId} • Workflow: {kyc.workflowId}
              </div>
            </div>
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

        {/* Body */}
        <div className="p-5 overflow-y-auto flex-1 space-y-5">
          {/* Top Info Grid */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-950/60 border border-slate-800 rounded-xl p-3.5 text-xs">
            <div>
              <span className="text-slate-500 block text-[10px]">Full Name</span>
              <span className="font-semibold text-white">{kyc.fullName}</span>
            </div>
            <div>
              <span className="text-slate-500 block text-[10px]">DOB &amp; Nationality</span>
              <span className="font-mono text-slate-300">
                {kyc.dateOfBirth} ({kyc.nationality})
              </span>
            </div>
            <div>
              <span className="text-slate-500 block text-[10px]">Document</span>
              <span className="font-mono text-slate-300">
                {kyc.documentType}: {kyc.documentNumber}
              </span>
            </div>
            <div>
              <span className="text-slate-500 block text-[10px]">Risk Score</span>
              <span className="font-mono font-bold text-amber-400">
                {kyc.riskScore !== undefined ? kyc.riskScore : 'Under Evaluation'}
              </span>
            </div>
          </div>

          {/* Photos */}
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-1.5">
              <Eye className="w-3.5 h-3.5 text-cyan-400" />
              <span>Uploaded KYC Verification Media</span>
            </h4>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-2.5">
                <span className="text-[10px] font-semibold text-slate-400 block mb-1">Front ID</span>
                <div className="aspect-[4/3] rounded-lg overflow-hidden bg-slate-950 border border-slate-800 flex items-center justify-center">
                  {kyc.frontDocumentUrl ? (
                    <img
                      src={kyc.frontDocumentUrl}
                      alt="Front"
                      className="w-full h-full object-cover cursor-pointer hover:scale-105 transition"
                      onClick={() => setZoomedImage({ url: kyc.frontDocumentUrl, title: 'Front ID' })}
                    />
                  ) : (
                    <span className="text-[10px] text-slate-600 p-2">No photo</span>
                  )}
                </div>
              </div>

              <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-2.5">
                <span className="text-[10px] font-semibold text-slate-400 block mb-1">Back ID</span>
                <div className="aspect-[4/3] rounded-lg overflow-hidden bg-slate-950 border border-slate-800 flex items-center justify-center">
                  {kyc.backDocumentUrl ? (
                    <img
                      src={kyc.backDocumentUrl}
                      alt="Back"
                      className="w-full h-full object-cover cursor-pointer hover:scale-105 transition"
                      onClick={() => setZoomedImage({ url: kyc.backDocumentUrl || '', title: 'Back ID' })}
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-[10px] text-slate-600">
                      Optional Back
                    </div>
                  )}
                </div>
              </div>

              <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-2.5">
                <span className="text-[10px] font-semibold text-slate-400 block mb-1">Live Selfie</span>
                <div className="aspect-[4/3] rounded-lg overflow-hidden bg-slate-950 border border-slate-800 flex items-center justify-center">
                  {kyc.selfieUrl ? (
                    <img
                      src={kyc.selfieUrl}
                      alt="Selfie"
                      className="w-full h-full object-cover cursor-pointer hover:scale-105 transition"
                      onClick={() => setZoomedImage({ url: kyc.selfieUrl, title: 'Selfie' })}
                    />
                  ) : (
                    <span className="text-[10px] text-slate-600 p-2">No selfie</span>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Compliance Review Input */}
          <div className="bg-slate-950/60 border border-slate-800 rounded-xl p-3.5 text-xs space-y-2.5">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-1.5">
              <Activity className="w-3.5 h-3.5 text-cyan-400" />
              <span>Compliance Review Decision Rationale</span>
            </h4>
            <div>
              <label className="block text-[11px] font-medium text-slate-400 mb-1">
                Officer Notes &amp; Audit Rationale
              </label>
              <textarea
                rows={3}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:ring-1 focus:ring-cyan-500 resize-none font-sans"
              />
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="px-5 py-3 border-t border-slate-800 bg-slate-950/70 flex items-center justify-between shrink-0">
          <div className="text-xs text-slate-400">
            Reviewer: <span className="font-mono font-semibold text-cyan-300">{reviewerId}</span>
          </div>

          <div className="flex items-center space-x-2.5">
            <button
              type="button"
              disabled={isSubmitting}
              onClick={() => handleAction('REJECT')}
              className="px-3.5 py-1.5 rounded-lg border border-rose-500/40 hover:bg-rose-500/20 text-rose-300 text-xs font-semibold flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <ShieldX className="w-3.5 h-3.5" />
              <span>Reject KYC</span>
            </button>

            <button
              type="button"
              disabled={isSubmitting}
              onClick={() => handleAction('APPROVE')}
              className="px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold shadow-xs flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <ShieldCheck className="w-4 h-4" />
              <span>{isSubmitting ? 'Submitting Signal...' : 'Manual Approve KYC'}</span>
            </button>
          </div>
        </div>
      </div>

      {zoomedImage && (
        <div
          className="fixed inset-0 z-60 bg-black/90 flex flex-col items-center justify-center p-4 cursor-pointer"
          onClick={() => setZoomedImage(null)}
        >
          <div className="max-w-3xl max-h-[85vh] bg-slate-900 border border-slate-700 rounded-xl overflow-hidden p-2">
            <img src={zoomedImage.url} alt={zoomedImage.title} className="max-h-[75vh] w-auto object-contain rounded-lg" />
          </div>
        </div>
      )}
    </div>
  );
};
