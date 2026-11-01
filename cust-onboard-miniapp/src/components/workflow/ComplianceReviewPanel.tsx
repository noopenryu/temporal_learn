import React, { useState } from 'react';
import { ShieldCheck, ShieldX, UserCheck } from 'lucide-react';
import { ManualReviewDecision } from '../../types/onboarding';

interface ComplianceReviewPanelProps {
  onDecision: (decision: ManualReviewDecision) => Promise<void>;
  isLoading: boolean;
}

export const ComplianceReviewPanel: React.FC<ComplianceReviewPanelProps> = ({
  onDecision,
  isLoading,
}) => {
  const [reviewerId, setReviewerId] = useState('OFFICER-COMPLIANCE-01');
  const [notes, setNotes] = useState('Identity and PEP checks verified against internal watchlist.');

  const handleAction = (approved: boolean) => {
    onDecision({
      reviewer_id: reviewerId.trim(),
      approved,
      notes: notes.trim(),
    });
  };

  return (
    <div className="bg-white border-2 border-purple-200 rounded-2xl p-6 shadow-md space-y-4">
      <div className="flex items-start space-x-3">
        <div className="p-2 rounded-xl bg-purple-100 text-purple-700 shrink-0">
          <UserCheck className="w-6 h-6" />
        </div>
        <div>
          <div className="flex items-center space-x-2">
            <h3 className="text-base font-bold text-slate-900">Compliance Officer Review Signal Trigger</h3>
            <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-purple-100 text-purple-800 border border-purple-200">
              SIMULATED ACTION
            </span>
          </div>
          <p className="text-xs text-slate-600 mt-0.5">
            This application triggered enhanced due diligence and is waiting for a compliance sign-off signal in Temporal. You can simulate the officer decision below or perform it in the Staff Console.
          </p>
        </div>
      </div>

      <div className="space-y-3 pt-1">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">Reviewer ID</label>
            <input
              type="text"
              required
              value={reviewerId}
              onChange={(e) => setReviewerId(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-xs text-slate-900 focus:bg-white focus:border-blue-600 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">Compliance Notes</label>
            <input
              type="text"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-xs text-slate-900 focus:bg-white focus:border-blue-600"
            />
          </div>
        </div>

        <div className="pt-2 flex items-center justify-end space-x-3">
          <button
            type="button"
            disabled={isLoading}
            onClick={() => handleAction(false)}
            className="px-4 py-2 rounded-xl border border-rose-200 hover:bg-rose-50 text-rose-700 text-xs font-semibold flex items-center space-x-1.5 transition disabled:opacity-50"
          >
            <ShieldX className="w-4 h-4 text-rose-600" />
            <span>Signal Reject</span>
          </button>

          <button
            type="button"
            disabled={isLoading}
            onClick={() => handleAction(true)}
            className="px-5 py-2 rounded-xl bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold flex items-center space-x-1.5 transition shadow-xs disabled:opacity-50"
          >
            <ShieldCheck className="w-4 h-4" />
            <span>Signal Approve (Release Sagas)</span>
          </button>
        </div>
      </div>
    </div>
  );
};
