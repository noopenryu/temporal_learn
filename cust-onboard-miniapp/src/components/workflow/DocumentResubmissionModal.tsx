import React, { useState } from 'react';
import { FileWarning, Send, Sparkles } from 'lucide-react';
import { DocumentType, IdentityDocuments } from '../../types/onboarding';

interface DocumentResubmissionModalProps {
  initialDocs: IdentityDocuments;
  rejectionReason?: string;
  onResubmit: (docs: Partial<IdentityDocuments>) => Promise<void>;
  isLoading: boolean;
}

export const DocumentResubmissionModal: React.FC<DocumentResubmissionModalProps> = ({
  initialDocs,
  rejectionReason,
  onResubmit,
  isLoading,
}) => {
  const [docType, setDocType] = useState<DocumentType>(
    (initialDocs?.document_type as DocumentType) || 'NATIONAL_ID'
  );
  const [docNumber, setDocNumber] = useState(initialDocs?.document_number || '');
  const [issuingCountry, setIssuingCountry] = useState(initialDocs?.issuing_country || 'VN');
  const [frontUrl, setFrontUrl] = useState(
    'https://images.unsplash.com/photo-1578632767115-351597cf2477?w=600&auto=format&fit=crop&q=80'
  );
  const [backUrl] = useState(initialDocs?.back_document_url || '');
  const [selfieUrl, setSelfieUrl] = useState(
    'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=600&auto=format&fit=crop&q=80'
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onResubmit({
      document_type: docType,
      document_number: docNumber.trim(),
      issuing_country: issuingCountry.trim().toUpperCase(),
      front_document_url: frontUrl.trim(),
      back_document_url: backUrl.trim() || undefined,
      selfie_url: selfieUrl.trim(),
    });
  };

  const applyClearPreset = () => {
    setFrontUrl('https://images.unsplash.com/photo-1578632767115-351597cf2477?w=600&auto=format&fit=crop&q=80');
    setSelfieUrl('https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=600&auto=format&fit=crop&q=80');
  };

  return (
    <div className="bg-white border-2 border-amber-300 rounded-2xl p-6 shadow-md">
      <div className="flex items-start space-x-3 mb-4">
        <div className="p-2 rounded-xl bg-amber-100 text-amber-700 shrink-0">
          <FileWarning className="w-6 h-6" />
        </div>
        <div>
          <h3 className="text-base font-bold text-slate-900">Action Required: Resubmit Clear Identity Documents</h3>
          <p className="text-xs text-amber-900 mt-0.5">
            {rejectionReason || 'Document or selfie image was flagged as blurry or unreadable by automated verification.'}
          </p>
        </div>
      </div>

      <div className="mb-4 p-3 rounded-xl bg-amber-50 border border-amber-200 flex items-center justify-between text-xs">
        <span className="text-amber-900">
          The Temporal saga is paused in <code className="text-amber-800 font-bold font-mono">WAITING_UPDATED_DOCS</code> state waiting for your signal.
        </span>
        <button
          type="button"
          onClick={applyClearPreset}
          className="px-2.5 py-1 rounded-lg bg-white border border-amber-300 text-amber-800 hover:bg-amber-100 font-semibold flex items-center space-x-1 shrink-0 ml-2"
        >
          <Sparkles className="w-3 h-3 text-amber-600" />
          <span>Fill Clear Docs</span>
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">Doc Type</label>
            <select
              value={docType}
              onChange={(e) => setDocType(e.target.value as DocumentType)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-xs text-slate-900 focus:bg-white focus:border-blue-600"
            >
              <option value="NATIONAL_ID">National ID (CCCD)</option>
              <option value="PASSPORT">Passport</option>
              <option value="DRIVING_LICENSE">Driver License</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">Doc Number</label>
            <input
              type="text"
              required
              value={docNumber}
              onChange={(e) => setDocNumber(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-xs text-slate-900 font-mono focus:bg-white focus:border-blue-600"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">Country</label>
            <input
              type="text"
              required
              maxLength={2}
              value={issuingCountry}
              onChange={(e) => setIssuingCountry(e.target.value.toUpperCase())}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-xs text-slate-900 font-mono focus:bg-white focus:border-blue-600"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="p-3 rounded-xl border border-slate-200 bg-slate-50 space-y-2">
            <span className="text-xs font-semibold text-slate-800 block">Replacement Front ID</span>
            <div className="aspect-16/9 rounded-lg overflow-hidden bg-white border border-slate-200">
              <img src={frontUrl} alt="Replacement Front" className="w-full h-full object-cover" />
            </div>
          </div>

          <div className="p-3 rounded-xl border border-slate-200 bg-slate-50 space-y-2">
            <span className="text-xs font-semibold text-slate-800 block">Replacement Liveness Selfie</span>
            <div className="aspect-16/9 rounded-lg overflow-hidden bg-white border border-slate-200">
              <img src={selfieUrl} alt="Replacement Selfie" className="w-full h-full object-cover" />
            </div>
          </div>
        </div>

        <div className="pt-2 flex justify-end">
          <button
            type="submit"
            disabled={isLoading}
            className="px-6 py-2.5 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-bold text-xs flex items-center space-x-1.5 transition shadow-xs disabled:opacity-50"
          >
            <Send className="w-4 h-4" />
            <span>{isLoading ? 'Submitting Signal...' : 'Send Resubmission Signal to Temporal'}</span>
          </button>
        </div>
      </form>
    </div>
  );
};
