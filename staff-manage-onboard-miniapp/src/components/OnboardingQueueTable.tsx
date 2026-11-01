import React from 'react';
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  XCircle,
  FileText,
  User,
  ArrowRight,
  FileCheck,
  ShieldAlert,
} from 'lucide-react';
import { ApplicationDetailResponse } from '../types/staff';

interface OnboardingQueueTableProps {
  items: ApplicationDetailResponse[];
  isLoading: boolean;
  onSelectApplication: (item: ApplicationDetailResponse) => void;
  onSeedFirst: () => void;
}

export const OnboardingQueueTable: React.FC<OnboardingQueueTableProps> = ({
  items,
  isLoading,
  onSelectApplication,
  onSeedFirst,
}) => {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'REQUIRES_MANUAL_REVIEW':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/10 text-amber-300 border border-amber-500/30">
            <AlertTriangle className="w-3 h-3 text-amber-400 shrink-0" />
            <span>Manual Review Required</span>
          </span>
        );
      case 'APPROVED':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-300 border border-emerald-500/30">
            <CheckCircle2 className="w-3 h-3 text-emerald-400 shrink-0" />
            <span>Approved</span>
          </span>
        );
      case 'PROCESSING':
      case 'SUBMITTED':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-cyan-500/10 text-cyan-300 border border-cyan-500/30">
            <Clock className="w-3 h-3 text-cyan-400 animate-spin shrink-0" />
            <span>Processing</span>
          </span>
        );
      case 'WAITING_UPDATED_DOCS':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-purple-500/10 text-purple-300 border border-purple-500/30">
            <FileText className="w-3 h-3 text-purple-400 shrink-0" />
            <span>Waiting Customer Docs</span>
          </span>
        );
      case 'REJECTED':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-rose-500/10 text-rose-300 border border-rose-500/30">
            <XCircle className="w-3 h-3 text-rose-400 shrink-0" />
            <span>Rejected</span>
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono bg-slate-800 text-slate-400">
            {status}
          </span>
        );
    }
  };

  if (isLoading && items.length === 0) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center">
        <div className="w-7 h-7 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
        <p className="text-xs text-slate-400">Syncing active onboarding cases with ledger...</p>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="bg-slate-900/60 border border-dashed border-slate-800 rounded-xl p-12 text-center">
        <div className="w-12 h-12 rounded-xl bg-slate-800/80 text-slate-400 flex items-center justify-center mx-auto mb-3 border border-slate-700/60">
          <User className="w-5 h-5 text-slate-300" />
        </div>
        <h3 className="text-sm font-semibold text-white mb-1">No Applications In Active Queue</h3>
        <p className="text-xs text-slate-400 max-w-sm mx-auto mb-4 leading-relaxed">
          There are currently no cases matching your queue criteria. Dispatch a synthetic compliance case to test manual verification and Temporal signals.
        </p>
        <button
          type="button"
          onClick={onSeedFirst}
          className="px-4 py-2 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold shadow-xs transition inline-flex items-center space-x-1.5"
        >
          <ShieldAlert className="w-3.5 h-3.5 text-cyan-200" />
          <span>Seed Manual Review Case Now</span>
        </button>
      </div>
    );
  }

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xs">
      {/* Table Subheader summary */}
      <div className="px-4 py-2.5 bg-slate-950/70 border-b border-slate-800 flex items-center justify-between text-xs text-slate-400">
        <span className="font-medium text-slate-300">
          Showing <span className="text-white font-mono font-semibold">{items.length}</span> cases in onboarding queue
        </span>
        <span className="text-[11px] font-mono text-slate-500">
          Sort: Most Recent First • 10s Live Polling
        </span>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-950/40 text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              <th className="py-3 px-4 font-mono">Status</th>
              <th className="py-3 px-4">Applicant / Customer</th>
              <th className="py-3 px-4">Requested Tier</th>
              <th className="py-3 px-4">Current Step</th>
              <th className="py-3 px-4">Officer Overrides</th>
              <th className="py-3 px-4 text-right font-mono">Timestamp</th>
              <th className="py-3 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60 font-sans">
            {items.map((item) => {
              const app = item.application;
              const fullName =
                `${app.personal_info?.first_name || ''} ${app.personal_info?.last_name || ''}`.trim() ||
                'Pending Information';
              const isManualReview = app.status === 'REQUIRES_MANUAL_REVIEW';

              return (
                <tr
                  key={app.id}
                  className={`hover:bg-slate-800/40 transition group ${
                    isManualReview ? 'bg-amber-500/[0.03]' : ''
                  }`}
                >
                  {/* Status Badge */}
                  <td className="py-3 px-4 whitespace-nowrap">
                    {getStatusBadge(app.status)}
                  </td>

                  {/* Customer Info */}
                  <td className="py-3 px-4">
                    <div className="font-semibold text-white group-hover:text-cyan-300 transition flex items-center space-x-1.5">
                      <span>{fullName}</span>
                      {app.personal_info?.nationality && (
                        <span className="text-[10px] px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 font-mono">
                          {app.personal_info.nationality}
                        </span>
                      )}
                    </div>
                    <div className="text-[11px] text-slate-400 font-mono flex items-center space-x-2 mt-0.5 tabular-nums">
                      <span className="text-slate-300">{app.customer_id}</span>
                      <span className="text-slate-600">•</span>
                      <span>{app.personal_info?.phone_number || 'No phone'}</span>
                    </div>
                  </td>

                  {/* Tier */}
                  <td className="py-3 px-4 whitespace-nowrap">
                    <span className="font-mono text-[11px] font-semibold text-slate-300">
                      {app.wallet_tier_requested || 'TIER_1_BASIC'}
                    </span>
                    {app.approved_tier && (
                      <div className="text-[10px] text-emerald-400 font-medium">
                        Approved: {app.approved_tier}
                      </div>
                    )}
                  </td>

                  {/* Current Step */}
                  <td className="py-3 px-4 whitespace-nowrap">
                    <span className="text-[11px] font-mono text-slate-400 bg-slate-950/80 px-2 py-0.5 rounded border border-slate-800">
                      {app.current_step || 'initial'}
                    </span>
                  </td>

                  {/* Confirmed Data Indicator */}
                  <td className="py-3 px-4 whitespace-nowrap">
                    {app.confirmed_data ? (
                      <span className="inline-flex items-center space-x-1 text-[11px] font-medium text-indigo-300 bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 rounded-md">
                        <FileCheck className="w-3 h-3 text-indigo-400" />
                        <span>Confirmed Data</span>
                      </span>
                    ) : (
                      <span className="text-[11px] text-slate-600 italic">Unconfirmed</span>
                    )}
                  </td>

                  {/* Date (Right-aligned tabular) */}
                  <td className="py-3 px-4 whitespace-nowrap text-slate-400 text-[11px] font-mono text-right tabular-nums">
                    {new Date(app.updated_at || app.created_at).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                      second: '2-digit',
                    })}
                  </td>

                  {/* Action Button */}
                  <td className="py-3 px-4 text-right whitespace-nowrap">
                    <button
                      type="button"
                      onClick={() => onSelectApplication(item)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-semibold inline-flex items-center space-x-1.5 transition ${
                        isManualReview
                          ? 'bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold shadow-xs'
                          : 'bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700'
                      }`}
                    >
                      <span>{isManualReview ? 'Review & Confirm' : 'View Detail'}</span>
                      <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
