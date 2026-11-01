import React from 'react';
import { AlertTriangle, CheckCircle2, Clock, XCircle, ArrowRight, ShieldAlert, Sparkles } from 'lucide-react';
import { KycApplicationDto } from '../types/staff';

interface KycQueueTableProps {
  items: KycApplicationDto[];
  isLoading: boolean;
  onSelectKyc: (item: KycApplicationDto) => void;
  onSeedKyc: () => void;
}

export const KycQueueTable: React.FC<KycQueueTableProps> = ({
  items,
  isLoading,
  onSelectKyc,
  onSeedKyc,
}) => {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'REQUIRES_MANUAL_REVIEW':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/10 text-amber-300 border border-amber-500/30">
            <AlertTriangle className="w-3 h-3 text-amber-400 shrink-0" />
            <span>Review Pending</span>
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
      case 'PENDING':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-cyan-500/10 text-cyan-300 border border-cyan-500/30">
            <Clock className="w-3 h-3 text-cyan-400 animate-spin shrink-0" />
            <span>Processing</span>
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
        <p className="text-xs text-slate-400">Loading KYC microservice records...</p>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="bg-slate-900/60 border border-dashed border-slate-800 rounded-xl p-12 text-center">
        <div className="w-12 h-12 rounded-xl bg-slate-800/80 text-slate-400 flex items-center justify-center mx-auto mb-3 border border-slate-700/60">
          <ShieldAlert className="w-5 h-5 text-slate-300" />
        </div>
        <h3 className="text-sm font-semibold text-white mb-1">No KYC Applications In Queue</h3>
        <p className="text-xs text-slate-400 max-w-sm mx-auto mb-4 leading-relaxed">
          There are no KYC microservice records currently matching. Dispatch a synthetic demo KYC workflow to test officer evaluation.
        </p>
        <button
          type="button"
          onClick={onSeedKyc}
          className="px-4 py-2 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold shadow-xs transition inline-flex items-center space-x-1.5"
        >
          <Sparkles className="w-3.5 h-3.5 text-cyan-200" />
          <span>Dispatch Test KYC Workflow</span>
        </button>
      </div>
    );
  }

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xs">
      {/* Subheader */}
      <div className="px-4 py-2.5 bg-slate-950/70 border-b border-slate-800 flex items-center justify-between text-xs text-slate-400">
        <span className="font-medium text-slate-300">
          Showing <span className="text-white font-mono font-semibold">{items.length}</span> KYC cases in queue
        </span>
        <span className="text-[11px] font-mono text-slate-500">
          Biometric &amp; Sanctions Verification Engine
        </span>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-950/40 text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              <th className="py-3 px-4 font-mono">Status</th>
              <th className="py-3 px-4">KYC Applicant</th>
              <th className="py-3 px-4">Document Details</th>
              <th className="py-3 px-4">Risk Evaluation</th>
              <th className="py-3 px-4 font-mono">Application ID</th>
              <th className="py-3 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60 font-sans">
            {items.map((item) => {
              const isReview = item.status === 'REQUIRES_MANUAL_REVIEW';
              return (
                <tr
                  key={item.applicationId}
                  className={`hover:bg-slate-800/40 transition group ${
                    isReview ? 'bg-amber-500/[0.03]' : ''
                  }`}
                >
                  <td className="py-3 px-4 whitespace-nowrap">
                    {getStatusBadge(item.status)}
                  </td>
                  <td className="py-3 px-4">
                    <div className="font-semibold text-white group-hover:text-cyan-300 transition flex items-center space-x-1.5">
                      <span>{item.fullName}</span>
                      <span className="text-[10px] px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 font-mono">
                        {item.nationality}
                      </span>
                    </div>
                    <div className="text-[11px] text-slate-400 font-mono mt-0.5 tabular-nums">
                      <span className="text-slate-300">{item.customerId}</span>
                      <span className="text-slate-600 mx-1.5">•</span>
                      <span>{item.phoneNumber}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap">
                    <div className="font-mono text-slate-300 text-[11px]">
                      {item.documentType}: {item.documentNumber}
                    </div>
                    <div className="text-[10px] text-slate-500 font-mono mt-0.5">DOB: {item.dateOfBirth}</div>
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap">
                    {item.riskScore !== undefined ? (
                      <div className="flex items-center space-x-2">
                        <span
                          className={`font-mono font-bold text-[11px] px-2 py-0.5 rounded ${
                            item.riskScore > 60
                              ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                              : item.riskScore > 30
                              ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                              : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          }`}
                        >
                          Score: {item.riskScore}
                        </span>
                      </div>
                    ) : (
                      <span className="text-slate-500 italic text-[11px]">Evaluating</span>
                    )}
                  </td>
                  <td className="py-3 px-4 whitespace-nowrap font-mono text-[11px] text-slate-400 tabular-nums">
                    {item.applicationId}
                  </td>
                  <td className="py-3 px-4 text-right whitespace-nowrap">
                    <button
                      type="button"
                      onClick={() => onSelectKyc(item)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-semibold inline-flex items-center space-x-1.5 transition ${
                        isReview
                          ? 'bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold shadow-xs'
                          : 'bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700'
                      }`}
                    >
                      <span>{isReview ? 'Review KYC' : 'View'}</span>
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
